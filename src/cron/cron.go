package cron

import (
	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
	"gitlab-mr-autocloser/src/config"
	"time"
)

type AsyncCronJob interface {
	StartAsyncCronJob(f func())
}

type cronJob struct {
	cfg *config.AutoCloserConfig
	log *logrus.Logger
}

func NewAsyncCronJob(cfg *config.AutoCloserConfig, log *logrus.Logger) AsyncCronJob {
	j := cronJob{
		cfg: cfg,
		log: log,
	}

	return &j
}

func (cj *cronJob) StartAsyncCronJob(f func()) {
	cj.log.Infof("Creating CRON task with schedule %s", cj.cfg.CronSchedule)

	s := gocron.NewScheduler(time.UTC)
	_, err := s.Cron(cj.cfg.CronSchedule).Do(f)

	if err != nil {
		cj.log.Errorf("CRON task was not created: %s\n", err)
	} else {
		s.StartAsync()
	}
}
