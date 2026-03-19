package clock

import (
	"time"
)

// テストしやすくするため現在時刻を直接使わず、差し替え可能にする
type Clocker interface {
	Now() time.Time
}

type RealClocker struct{}

func (r RealClocker) Now() time.Time {
	return time.Now()
}

// テスト用
type FixedClocker struct{}

func (fc FixedClocker) Now() time.Time {
	// 常に同じ時間を返す
	return time.Date(2022, 5, 10, 12, 34, 56, 0, time.UTC)
}
