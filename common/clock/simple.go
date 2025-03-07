package clock

import (
	"sync/atomic"
	"time"
)

// SimpleClock 是一个简单的时钟结构体，使用原子操作来设置和获取时间。
// 它包含一个 atomic.Pointer[time.Time] 类型的字段 v，用于原子性地存储时间指针。
type SimpleClock struct {
	v atomic.Pointer[time.Time]
}

// NewSimpleClock 创建并返回一个新的 SimpleClock 实例。
func NewSimpleClock() *SimpleClock {
	return &SimpleClock{}
}

// SetTime 根据给定的 Unix 时间戳 u 设置时钟的时间。
// 参数:
//
//	u - Unix 时间戳，表示自1970年1月1日UTC以来的秒数。
func (c *SimpleClock) SetTime(u uint64) {
	t := time.Unix(int64(u), 0)
	c.v.Store(&t)
}

// Set 直接设置时钟的时间为给定的 time.Time 值。
// 参数:
//
//	v - 要设置的时间值。
func (c *SimpleClock) Set(v time.Time) {
	c.v.Store(&v)
}

// Now 返回当前时钟的时间。
// 如果时钟未被设置，返回 Unix 纪元时间（1970年1月1日UTC）。
// 返回值:
//
//	time.Time - 当前时钟的时间。
func (c *SimpleClock) Now() time.Time {
	v := c.v.Load()
	if v == nil {
		return time.Unix(0, 0)
	}
	return *v
}
