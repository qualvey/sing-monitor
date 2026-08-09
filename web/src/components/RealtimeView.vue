<template>
  <div class="space-y-6">
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <!-- 实时速率图表 -->
      <div class="lg:col-span-2 bg-[#131b2e] border border-slate-800 rounded-2xl p-4">
        <div class="flex items-center justify-between mb-2">
          <h3 class="text-base font-bold">⚡ 实时速率 (bytes/s) — {{ chartTargetName }}</h3>
          <select v-model="chartTarget" @change="chartSeries = []"
            class="px-2 py-1 text-xs bg-slate-900 border border-slate-700 rounded-lg focus:outline-none focus:border-indigo-500 max-w-[180px]">
            <option value="all">全部用户（合计）</option>
            <option v-for="u in users" :key="u.name" :value="u.name">{{ u.name }}</option>
          </select>
        </div>
        <div ref="chartEl" class="h-64"></div>
      </div>
      <div class="bg-[#131b2e] border border-slate-800 rounded-2xl p-4 space-y-3">
        <div class="flex items-center justify-between">
          <div class="text-xs text-slate-400 uppercase">采集灵敏度</div>
          <select v-model="intervalMs" @change="sendInterval"
            class="px-2 py-1 text-xs bg-slate-900 border border-slate-700 rounded-lg focus:outline-none focus:border-indigo-500">
            <option :value="1000">1 秒（最快）</option>
            <option :value="2000">2 秒</option>
            <option :value="5000">5 秒</option>
            <option :value="10000">10 秒</option>
          </select>
        </div>
        <div class="text-[11px] text-slate-500">实时监控页打开时自动切高频采集，关闭页面自动恢复默认</div>
        <div>
          <div class="text-xs text-slate-400 uppercase">上行速率</div>
          <div class="text-2xl font-bold text-indigo-300">{{ fmtBytes(globalUp) }}/s</div>
        </div>
        <div>
          <div class="text-xs text-slate-400 uppercase">下行速率</div>
          <div class="text-2xl font-bold text-cyan-300">{{ fmtBytes(globalDown) }}/s</div>
        </div>
        <div>
          <div class="text-xs text-slate-400 uppercase">连接状态</div>
          <div class="flex items-center gap-2 mt-1">
            <span :class="connected ? 'bg-emerald-400' : 'bg-red-500'" class="w-2.5 h-2.5 rounded-full animate-pulse"></span>
            <span class="text-sm">{{ connected ? '已连接 (实时推送)' : '连接断开，重连中...' }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 用户实时状态表 -->
    <div class="bg-[#131b2e] border border-slate-800 rounded-2xl overflow-x-auto">
      <table class="w-full text-left text-xs">
        <thead class="bg-slate-900/80 text-slate-400 uppercase font-semibold border-b border-slate-800">
          <tr>
            <SortableTh label="用户" :active="sortKey === 'name'" :dir="sortDir" @sort="toggleSort('name')" />
            <SortableTh label="上行速率" :active="sortKey === 'uplink'" :dir="sortDir" @sort="toggleSort('uplink')" />
            <SortableTh label="下行速率" :active="sortKey === 'downlink'" :dir="sortDir" @sort="toggleSort('downlink')" />
            <SortableTh label="状态" :active="sortKey === 'online'" :dir="sortDir" @sort="toggleSort('online')" />
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-800/60">
          <tr v-for="u in sortedUsers" :key="u.name" @click="chartTarget = u.name; chartSeries.length = 0"
            :class="['hover:bg-slate-800/40 cursor-pointer transition-colors', chartTarget === u.name ? 'bg-indigo-500/5' : '']">
            <td class="py-3 px-4 font-semibold">{{ u.name }}</td>
            <td class="py-3 px-4 font-mono text-indigo-400">{{ fmtBytes(u.uplink) }}/s</td>
            <td class="py-3 px-4 font-mono text-cyan-400">{{ fmtBytes(u.downlink) }}/s</td>
            <td class="py-3 px-4">
              <span :class="u.online ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-slate-800 text-slate-400 border-slate-700',
                'inline-flex px-2 py-0.5 rounded-full text-[11px] font-semibold border'">
                {{ u.online ? '在线' : '离线' }}
              </span>
            </td>
          </tr>
          <tr v-if="users.length === 0"><td colspan="4" class="py-12 text-center text-slate-500">等待实时数据...</td></tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import * as echarts from 'echarts'
import { getToken, fmtBytes } from '../api'
import SortableTh from './SortableTh.vue'

const chartEl = ref(null)
const connected = ref(false)
const globalUp = ref(0)
const globalDown = ref(0)
const users = ref([])
const intervalMs = ref(2000)
const chartTarget = ref('all')
const chartSeries = [] // [时间, up, down] 三元组

const chartTargetName = computed(() =>
  chartTarget.value === 'all' ? '全部用户' : chartTarget.value)

// 按当前图表目标取速率
function targetRates() {
  if (chartTarget.value === 'all') {
    return [globalUp.value, globalDown.value]
  }
  const u = users.value.find(x => x.name === chartTarget.value)
  return u ? [u.uplink, u.downlink] : [0, 0]
}

function sendInterval() {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ action: 'set_interval', interval_ms: intervalMs.value }))
  }
}

// 表头排序（默认速率降序）
const sortKey = ref('downlink')
const sortDir = ref('desc')

function toggleSort(k) {
  if (sortKey.value === k) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = k
    sortDir.value = 'asc'
  }
}

const sortedUsers = computed(() => {
  const arr = [...users.value]
  const k = sortKey.value
  const dir = sortDir.value === 'asc' ? 1 : -1
  arr.sort((a, b) => {
    const va = a[k]
    const vb = b[k]
    if (typeof va === 'string' && typeof vb === 'string') return va.localeCompare(vb, 'zh') * dir
    return (va - vb) * dir
  })
  return arr
})

let ws = null
let chart = null

function connect() {
  // 相对路径 WebSocket：支持 nginx 子路径反代
  const url = new URL('api/v1/ws/rt', window.location.href)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.search = getToken() ? `?token=${getToken()}` : ''
  ws = new WebSocket(url.href)
  ws.onopen = () => { connected.value = true; sendInterval() }
  ws.onmessage = (e) => {
    try {
      const d = JSON.parse(e.data)
      globalUp.value = d.global?.uplink || 0
      globalDown.value = d.global?.downlink || 0
      users.value = d.users || []
      pushChart()
    } catch {}
  }
  ws.onclose = () => {
    connected.value = false
    setTimeout(connect, 2000)
  }
  ws.onerror = () => ws && ws.close()
}

function pushChart() {
  const [up, down] = targetRates()
  // x 轴必须用时间戳（epoch ms），time 轴才能正确渲染
  const t = Date.now()
  chartSeries.push([t, up, down])
  if (chartSeries.length > 60) { chartSeries.shift() }
  if (chart) {
    chart.setOption({
      series: [
        { name: '上行', data: chartSeries.map(p => [p[0], p[1]]) },
        { name: '下行', data: chartSeries.map(p => [p[0], p[2]]) },
      ],
    })
  }
}

// 速率友好格式化（bytes/s → KB/MB/GB per s）
function fmtRate(v) {
  if (v >= 1e9) return (v / 1e9).toFixed(2) + ' GB/s'
  if (v >= 1e6) return (v / 1e6).toFixed(2) + ' MB/s'
  if (v >= 1e3) return (v / 1e3).toFixed(1) + ' KB/s'
  return v + ' B/s'
}

onMounted(async () => {
  await nextTick()
  chart = echarts.init(chartEl.value)
  chart.setOption({
    tooltip: {
      trigger: 'axis',
      valueFormatter: fmtRate,
    },
    legend: { data: ['上行', '下行'], textStyle: { color: '#94a3b8' } },
    grid: { left: 70, right: 20, top: 30, bottom: 30 },
    xAxis: { type: 'time', axisLabel: { color: '#64748b' } },
    yAxis: {
      type: 'value',
      // 不从 0 开始，放大小流量波动；轴标签格式化单位
      scale: true,
      axisLabel: { color: '#64748b', formatter: fmtRate },
    },
    // 滚轮/拖拽缩放
    dataZoom: [{ type: 'inside' }, { type: 'slider', height: 14, bottom: 5 }],
    series: [
      { name: '上行', type: 'line', smooth: true, showSymbol: false, data: [], lineStyle: { color: '#818cf8' }, itemStyle: { color: '#818cf8' }, areaStyle: { opacity: 0.08 } },
      { name: '下行', type: 'line', smooth: true, showSymbol: false, data: [], lineStyle: { color: '#22d3ee' }, itemStyle: { color: '#22d3ee' }, areaStyle: { opacity: 0.08 } },
    ],
  })
  connect()
})

onBeforeUnmount(() => {
  if (ws) ws.close()
  if (chart) chart.dispose()
})
</script>
