package models

import (
	"testing"
	"time"
)

func mustTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestCurrentCycleWindow(t *testing.T) {
	anchor := mustTime("2026-08-01T00:00:00Z")

	cases := []struct {
		name  string
		user  User
		now   time.Time
		wantS string
		wantE string
	}{
		{
			name:  "窗口期内",
			user:  User{CycleStart: anchor, CycleDays: 30},
			now:   mustTime("2026-08-15T12:00:00Z"),
			wantS: "2026-08-01T00:00:00Z",
			wantE: "2026-08-31T00:00:00Z",
		},
		{
			name:  "滚动到第二周期",
			user:  User{CycleStart: anchor, CycleDays: 30},
			now:   mustTime("2026-09-01T00:00:00Z"),
			wantS: "2026-08-31T00:00:00Z",
			wantE: "2026-09-30T00:00:00Z",
		},
		{
			name:  "跨多个周期",
			user:  User{CycleStart: anchor, CycleDays: 30},
			now:   mustTime("2026-12-15T00:00:00Z"),
			wantS: "2026-11-29T00:00:00Z", // 08-01 + 3*30d = 10-30, +30d = 11-29
			wantE: "2026-12-29T00:00:00Z",
		},
		{
			name:  "零值CycleStart回退CreatedAt",
			user:  User{CreatedAt: anchor, CycleDays: 30},
			now:   mustTime("2026-08-15T00:00:00Z"),
			wantS: "2026-08-01T00:00:00Z",
			wantE: "2026-08-31T00:00:00Z",
		},
		{
			name:  "未到锚点",
			user:  User{CycleStart: anchor, CycleDays: 30},
			now:   mustTime("2026-07-01T00:00:00Z"),
			wantS: "2026-08-01T00:00:00Z",
			wantE: "2026-08-31T00:00:00Z",
		},
		{
			name:  "非30天周期",
			user:  User{CycleStart: anchor, CycleDays: 7},
			now:   mustTime("2026-08-10T00:00:00Z"),
			wantS: "2026-08-08T00:00:00Z", // 08-01 + 1*7d
			wantE: "2026-08-15T00:00:00Z",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotS, gotE := tc.user.CurrentCycleWindow(tc.now)
			wantS := mustTime(tc.wantS)
			wantE := mustTime(tc.wantE)
			if !gotS.Equal(wantS) || !gotE.Equal(wantE) {
				t.Errorf("got [%v, %v), want [%v, %v)", gotS, gotE, wantS, wantE)
			}
		})
	}
}
