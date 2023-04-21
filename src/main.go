package main

import (
	"flag"
	"gitlab-mr-autocloser/src/config"
	"gitlab-mr-autocloser/src/cron"
	"gitlab-mr-autocloser/src/gitlab"
	"gitlab-mr-autocloser/src/healthcheck"
	"gitlab-mr-autocloser/src/logger"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	verboseFlag := flag.Bool("verbose", false, "enable verbose output")
	configPathFlag := flag.String("config", "./config/config.yaml", "config file path")
	flag.Parse()

	log := *logger.SetupLogger(*verboseFlag)
	if *verboseFlag {
		log.Infof("Verbose mode is enabled...")
	}

	log.Infof("Loading config from %s ...", *configPathFlag)
	cfg := *config.LoadConfig(*configPathFlag, &log)

	healthcheck.NewListener(&cfg, &log).StartHandlers()

	task := func() {
		gitlab.NewMRCloser(&cfg, &log).ManageMergeRequests()
	}

	cron.NewAsyncCronJob(&cfg, &log).StartAsyncCronJob(task)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	exit := <-sigChan
	log.Infof("Stopped by signal: %v", exit)
}
