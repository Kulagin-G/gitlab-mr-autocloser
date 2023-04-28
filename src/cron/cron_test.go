package cron

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gitlab-mr-autocloser/src/config"
	"sync"
	"testing"
)

func TestCronJob_StartAsyncCronJob(t *testing.T) {
	t.Setenv("GITLAB_API_TOKEN", "def456")

	log := logrus.New()
	cfg := *config.LoadConfig("../../config/config.yaml", log)

	cfg.CronSchedule = "* * * * * "

	cj := NewAsyncCronJob(&cfg, log)

	var wg sync.WaitGroup

	wg.Add(1)

	counter := 0
	f := func() {
		counter++

		wg.Done()
	}

	cj.StartAsyncCronJob(f)

	wg.Wait()

	// check if the counter was incremented
	assert.NotEqual(t, 0, counter)
}
