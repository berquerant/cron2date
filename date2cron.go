package cron2date

import (
	"fmt"
	"time"
)

func Date2Cron(t time.Time) string {
	return fmt.Sprintf(
		"%d %d %d %d *",
		t.Minute(),
		t.Hour(),
		t.Day(),
		t.Month(),
	)
}
