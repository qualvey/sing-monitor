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

func ts(t time.Time) *time.Time { return &t }

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
			user:  User{CycleStart: ts(anchor), CycleDays: 30},
			now:   mustTime("2026-08-15T12:00:00Z"),
			wantS: "2026-08-01T00:00:00Z",
			wantE: "2026-08-31T00:00:00Z",
		},
		{
			name:  "滚动到第二周期",
			user:  User{CycleStart: ts(anchor), CycleDays: 30},
			now:   mustTime("2026-09-01T00:00:00Z"),
			wantS: "2026-08-31T00:00:00Z",
			wantE: "2026-09-30T00:00:00Z",
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
			user:  User{CycleStart: ts(anchor), CycleDays: 30},
			now:   mustTime("2026-07-01T00:00:00Z"),
			wantS: "2026-08-01T00:00:00Z",
			wantE: "2026-08-31T00:00:00Z",
		},
		{
			name:  "非30天周期",
			user:  User{CycleStart: ts(anchor), CycleDays: 7},
			now:   mustTime("2026-08-10T00:00:00Z"),
			wantS: "2026-08-08T00:00:00Z",
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

func TestCycleAnchorFallback(t *testing.T) {
	created := mustTime("2026-07-01T00:00:00Z")
	explicit := mustTime("2026-06-01T00:00:00Z")

	u1 := User{CreatedAt: created}
	if !u1.CycleAnchor().Equal(created) {
		t.Errorf("expected fallback to CreatedAt, got %v", u1.CycleAnchor())
	}
	u2 := User{CreatedAt: created, CycleStart: ts(explicit)}
	if !u2.CycleAnchor().Equal(explicit) {
		t.Errorf("expected explicit CycleStart, got %v", u2.CycleAnchor())
	}
}

func TestIsOverLimit(t *testing.T) {
	u := User{TrafficLimit: 100}
	if u.IsOverLimit(99) {
		t.Error("99 should not exceed 100")
	}
	if !u.IsOverLimit(101) {
		t.Error("101 should exceed 100")
	}
	u3 := User{TrafficLimit: 0}
	if u3.IsOverLimit(9999) {
		t.Error("limit 0 means unlimited")
	}
}
