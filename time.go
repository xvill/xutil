package xtools

import (
	"fmt"
	"strings"
	"time"
)

func TimeTrunc(s string, t time.Time) time.Time {
	y, m, d := t.Date()
	H, M, S := t.Clock()
	switch strings.ToLower(s) {
	case "month":
		return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	case "day":
		return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	case "hour":
		return time.Date(y, m, d, H, 0, 0, 0, time.UTC)
	case "minute":
		return time.Date(y, m, d, H, M, 0, 0, time.UTC)
	case "second":
		return time.Date(y, m, d, H, M, S, 0, time.UTC)
	case "week":
		return time.Date(y, m, d-int(t.Weekday()), 0, 0, 0, 0, time.UTC)

	}
	return t
}

func TimeRule(s string, t time.Time, sum int) (ret []time.Time) {
	t = TimeTrunc(s, t)
	c := t
	sign := 1
	if sum > 0 {
		sign = 1
	} else {
		sign = -1
	}

	for i := 0; i < sum*sign; i++ {
		ret = append(ret, c)
		switch strings.ToLower(s) {
		case "month":
			c = c.AddDate(0, 1*sign, 0)
		case "day":
			c = c.AddDate(0, 0, 1*sign)
		case "week":
			c = c.AddDate(0, 0, 7*sign)
		case "hour":
			c = c.Add(time.Hour * time.Duration(sign))
		case "minute":
			c = c.Add(time.Minute * time.Duration(sign))
		case "second":
			c = c.Add(time.Second * time.Duration(sign))
		}
	}
	return
}

// DayLastRange 上周/上月/昨天
func DayLastRange(f string) (ret []string) {
	y, m, d := time.Now().Date()
	t := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	var btime, etime, curr time.Time
	switch strings.ToLower(f) {
	case "week":
		etime = t.AddDate(0, 0, -int(t.Weekday()))
		btime = etime.AddDate(0, 0, -7)
	case "month":
		etime = time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
		btime = time.Date(y, m-1, 1, 0, 0, 0, 0, time.UTC)
	case "day":
		etime = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
		btime = time.Date(y, m, d-1, 0, 0, 0, 0, time.UTC)
	}
	curr = btime
	for curr.Before(etime) {
		ret = append(ret, curr.Format("20060102"))
		curr = curr.AddDate(0, 0, 1)
	}
	return
}

func TimeFormat(f string, t ...time.Time) (ret []string) {
	for _, v := range t {
		ret = append(ret, v.Format(f))
	}
	return
}

func TimeSplit(b, e time.Time, d time.Duration) (r []time.Time) {
	for c := b; c.Before(e); c = c.Add(d) {
		r = append(r, c)
	}
	return
}

func TimeExtend(t time.Time, s int, d time.Duration) (r []time.Time) {
	t = t.UTC().Truncate(d)
	b, e := t, t.Add(time.Duration(s)*d)
	if s < 0 {
		b, e = t.Add(time.Duration(s)*d), t
	}
	return TimeSplit(b, e, d)
}

func test() {
	x := TimeSplit(time.Now(), time.Now().AddDate(0, 0, 10), 24*time.Hour)
	y := TimeFormat("20060102", x...)
	fmt.Println(y)

	x = TimeExtend(time.Now(), 10, 24*time.Hour)
	y = TimeFormat("20060102", x...)
	fmt.Println(y)
}
