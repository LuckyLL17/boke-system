package service

import (
	"testing"
	"time"
)

// ChannelDashboard -> GetChannelDashboard -> GroupByDate -> buildDatePoints
func TestDailyListenersRemainDistinctFromPlayCount(t *testing.T) {
	from := time.Date(2026, 8, 23, 0, 0, 0, 0, time.Local)
	rows := []map[string]interface{}{
		{"date": from, "count": int64(5), "listeners": int64(2)},
	}
	points := buildDatePoints(rows, from, time.Date(2026, 8, 23, 23, 59, 59, 0, time.Local), "listeners")
	if len(points) != 1 || points[0].Count != 2 {
		panic("daily listeners were replaced by play count")
	}
}
