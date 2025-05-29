package config

// Config is the root structure for the YAML config file.
type Config struct {
	SnapraidBin string       `yaml:"snapraid_bin"`
	ConfigFile  string       `yaml:"config_file"`
	OutputDir   string       `yaml:"output_dir"`
	Thresholds  Thresholds   `yaml:"thresholds"`
	Steps       Steps        `yaml:"steps"`
	Scrub       ScrubOptions `yaml:"scrub"`
	Notify      Notify       `yaml:"notifications"`
}

// Thresholds define numeric limits before sync is blocked.
type Thresholds struct {
	Add     int `yaml:"add"`
	Remove  int `yaml:"remove"`
	Update  int `yaml:"update"`
	Copy    int `yaml:"copy"`
	Move    int `yaml:"move"`
	Restore int `yaml:"restore"`
}

// Steps define which SnapRAID subcommands to run.
type Steps struct {
	Touch bool `yaml:"touch"`
	Scrub bool `yaml:"scrub"`
	Smart bool `yaml:"smart"`
}

// ScrubOptions control the `scrub` command.
type ScrubOptions struct {
	Plan      int `yaml:"plan"`
	OlderThan int `yaml:"older_than"`
}

// Notify defines Slack notification options.
type Notify struct {
	SlackToken   string `yaml:"slack_token"`
	SlackChannel string `yaml:"slack_channel"`
}
