package cron2date_test

import (
	"testing"
	"time"

	"github.com/berquerant/cron2date"
	"github.com/robfig/cron"
	"github.com/stretchr/testify/assert"
)

func TestCron2Date(t *testing.T) {
	const (
		targetString = "2100-10-30 10:02:00"
		want         = "2 10 30 10 *"
	)
	target, err := time.Parse(time.DateTime, targetString)
	assert.Nil(t, err)
	got := cron2date.Date2Cron(target)
	assert.Equal(t, want, got)

	sched, err := cron.ParseStandard(got)
	assert.Nil(t, err)
	nextTime := sched.Next(target)
	assert.Equal(t, target.Add(time.Hour*24*365), nextTime)
}
