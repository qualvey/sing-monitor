package realtime

import (
	"sort"
	"sync"
	"time"
)

// Snapshot 实时推送的负载结构（对齐原系统 WebSocket 消息）
type Snapshot struct {
	Global struct {
		Uplink   int64 `json:"uplink"`
		Downlink int64 `json:"downlink"`
	} `json:"global"`
	Users []UserRt `json:"users"`
}

// UserRt 单个用户实时状态
type UserRt struct {
	Name      string `json:"name"`
	Uplink    int64  `json:"uplink"`
	Downlink int64  `json:"downlink"`
	Online    bool   `json:"online"`
}

type targetState struct {
	lastUp   int64 // 上次累计上行
	lastDown int64 // 上次累计下行
	lastSeen time.Time
	rateUp   int64 // bytes/s
	rateDown int64 // bytes/s
}

// Broadcaster 聚合采集器增量，周期性推送 WebSocket 快照
type Broadcaster struct {
	intervalMs   int
	onlineThresh time.Duration

	mu       sync.Mutex
	states   map[string]*targetState
	subs     map[chan Snapshot]struct{}
	stop     chan struct{}
	done     chan struct{}
}

func NewBroadcaster(intervalMs int, onlineThresholdSec int) *Broadcaster {
	if intervalMs <= 0 {
		intervalMs = 1000
	}
	if onlineThresholdSec <= 0 {
		onlineThresholdSec = 120
	}
	return &Broadcaster{
		intervalMs:   intervalMs,
		onlineThresh: time.Duration(onlineThresholdSec) * time.Second,
		states:       make(map[string]*targetState),
		subs:         make(map[chan Snapshot]struct{}),
		stop:         make(chan struct{}),
		done:         make(chan struct{}),
	}
}

func (b *Broadcaster) Start() {
	go func() {
		defer close(b.done)
		ticker := time.NewTicker(time.Duration(b.intervalMs) * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				b.publish()
			case <-b.stop:
				return
			}
		}
	}()
}

func (b *Broadcaster) Stop() {
	close(b.stop)
	<-b.done
}

// Submit 采集器每轮调用：name=target(tag)，up/down=自上次以来的增量
// 仅在有真实流量时刷新活跃时间戳（在线判定依据）
func (b *Broadcaster) Submit(name string, up, down int64) {
	if up == 0 && down == 0 {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	st, ok := b.states[name]
	if !ok {
		st = &targetState{lastSeen: now}
		b.states[name] = st
	}
	// 速率按最近两次采样的增量与时间差计算
	if !st.lastSeen.IsZero() {
		dt := now.Sub(st.lastSeen).Seconds()
		if dt > 0 {
			st.rateUp = int64(float64(up) / dt)
			st.rateDown = int64(float64(down) / dt)
		}
	}
	st.lastSeen = now
}

func (b *Broadcaster) publish() {
	b.mu.Lock()
	now := time.Now()
	snap := Snapshot{}
	for name, st := range b.states {
		u := UserRt{
			Name:      name,
			Uplink:    st.rateUp,
			Downlink:  st.rateDown,
			Online:    now.Sub(st.lastSeen) < b.onlineThresh,
		}
		snap.Users = append(snap.Users, u)
		snap.Global.Uplink += st.rateUp
		snap.Global.Downlink += st.rateDown
	}
	// 稳定排序：按总速率降序（活跃在前），同速率按名称升序，避免每次推送顺序跳动
	sort.SliceStable(snap.Users, func(i, j int) bool {
		ri := snap.Users[i].Uplink + snap.Users[i].Downlink
		rj := snap.Users[j].Uplink + snap.Users[j].Downlink
		if ri != rj {
			return ri > rj
		}
		return snap.Users[i].Name < snap.Users[j].Name
	})
	subs := make([]chan Snapshot, 0, len(b.subs))
	for ch := range b.subs {
		subs = append(subs, ch)
	}
	b.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- snap:
		default:
		}
	}
}

func (b *Broadcaster) Subscribe() chan Snapshot {
	ch := make(chan Snapshot, 8)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *Broadcaster) Unsubscribe(ch chan Snapshot) {
	b.mu.Lock()
	delete(b.subs, ch)
	b.mu.Unlock()
}
