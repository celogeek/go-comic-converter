package epuboptions

type Crop struct {
	Enabled            bool `yaml:"enabled"`
	Left               int  `yaml:"left"`
	Up                 int  `yaml:"up"`
	Right              int  `yaml:"right"`
	Bottom             int  `yaml:"bottom"`
	Limit              int  `yaml:"limit"`
	SkipIfLimitReached bool `yaml:"skip_if_limit_reached"`
}
