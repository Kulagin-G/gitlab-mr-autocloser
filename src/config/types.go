package config

type AutoCloserConfig struct {
	GitlabApiToken     string             `yaml:"gitlabApiToken"`
	GitlabBaseApiUrl   string             `yaml:"gitlabBaseApiUrl"`
	CronSchedule       string             `yaml:"schedule"`
	LabelHead          string             `yaml:"labelHead"`
	HealthcheckOptions HealthcheckOptions `yaml:"healthcheckOptions"`
	DefaultOptions     DefaultOptions     `yaml:"defaultOptions"`
	Projects           []ProjectConfigs   `yaml:"projects"`
}

type HealthcheckOptions struct {
	Host      string    `yaml:"host"`
	Port      int       `yaml:"port"`
	Liveness  Liveness  `yaml:"liveness"`
	Readiness Readiness `yaml:"readiness"`
}

type Liveness struct {
	Path      string `yaml:"path"`
	GorMaxNum int    `yaml:"gorMaxNum"`
}

type Readiness struct {
	Path              string `yaml:"path"`
	ResolveTimeoutSec int    `yaml:"resolveTimeoutSec"`
	UrlCheck          string `yaml:"urlCheck"`
}

type DefaultOptions struct {
	StaleMRAfterDays int `yaml:"staleMRAfterDays"`
	CloseMRAfterDays int `yaml:"closeMRAfterDays"`
}

type ProjectConfigs struct {
	Name            string          `yaml:"name"`
	OverrideOptions OverrideOptions `yaml:"overrideOptions"`
}

type OverrideOptions struct {
	StaleMRAfterDays int `yaml:"staleMRAfterDays"`
	CloseMRAfterDays int `yaml:"closeMRAfterDays"`
}
