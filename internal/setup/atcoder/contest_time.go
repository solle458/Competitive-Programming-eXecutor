package atcoder

import "time"

func (c Contest) InProgress(now time.Time) bool {
	start := time.Unix(c.StartEpochSecond, 0)
	end := start.Add(time.Duration(c.DurationSecond) * time.Second)
	return !now.Before(start) && now.Before(end)
}
