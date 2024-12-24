package epuboptions

type Crop struct {
	Enabled            bool `yaml:"enabled" json:"enabled"`
	Left               int  `yaml:"left" json:"left"`
	Up                 int  `yaml:"up" json:"up"`
	Right              int  `yaml:"right" json:"right"`
	Bottom             int  `yaml:"bottom" json:"bottom"`
	Limit              int  `yaml:"limit" json:"limit"`
	SkipIfLimitReached bool `yaml:"skip_if_limit_reached" json:"skip_if_limit_reached"`
}
