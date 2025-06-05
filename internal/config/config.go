package config

// Config is the root structure for the YAML config file.
type Config struct {
	SnapraidBin    string       `yaml:"snapraid_bin"`    // SnapraidBin is the path to the snapraid executable (e.g., /usr/bin/snapraid).
	SnapraidConfig string       `yaml:"snapraid_config"` // SnapraidConfig is the path to the snapraid configuration file used by the snapraid command.
	OutputDir      string       `yaml:"output_dir"`      // OutputDir is the directory where JSON result files will be written. Leave empty to disable.
	Thresholds     Thresholds   `yaml:"thresholds"`      // Thresholds defines numeric limits for file-change categories before blocking sync.
	Steps          Steps        `yaml:"steps"`           // Steps toggles which SnapRAID subcommands to run (touch, scrub, smart).
	Scrub          ScrubOptions `yaml:"scrub"`           // Scrub holds options for the "scrub" command (plan percentage and file age threshold).
	Notify         Notify       `yaml:"notifications"`   // Notify contains Slack notification settings (token and channel).
}

// WantsSlackNotification returns true if Slack notifications
// should be sent (i.e. not suppressed, and both token+channel are set).
func (c Config) WantsSlackNotification(noNotify bool) bool {
	return !noNotify &&
		c.Notify.SlackChannel != "" &&
		c.Notify.SlackToken != ""
}

// Thresholds define numeric limits before sync is blocked.
type Thresholds struct {
	Add     *int `yaml:"add"`     // Add is the maximum number of added files allowed before aborting sync. Set to –1 to disable.
	Remove  *int `yaml:"remove"`  // Remove is the maximum number of removed files allowed before aborting sync. Set to –1 to disable.
	Update  *int `yaml:"update"`  // Update is the maximum number of updated files allowed before aborting sync. Set to –1 to disable.
	Copy    *int `yaml:"copy"`    // Copy is the maximum number of copied files allowed before aborting sync. Set to –1 to disable.
	Move    *int `yaml:"move"`    // Move is the maximum number of moved files allowed before aborting sync. Set to –1 to disable.
	Restore *int `yaml:"restore"` // Restore is the maximum number of restored files allowed before aborting sync. Set to –1 to disable.
}

// Steps define which SnapRAID subcommands to run.
type Steps struct {
	Touch *bool `yaml:"touch"` // Touch enables the "snapraid touch" step before sync.
	Scrub *bool `yaml:"scrub"` // Scrub enables the "snapraid scrub" step after sync.
	Smart *bool `yaml:"smart"` // Smart enables the "snapraid smart" step after scrub.
}

// ScrubOptions control the `scrub` command.
type ScrubOptions struct {
	Plan      *int `yaml:"plan"`       // Plan is the percentage (0–100) used by "snapraid scrub".
	OlderThan *int `yaml:"older_than"` // OlderThan is the minimum file age in days for "snapraid scrub" to include.
}

// Notify defines Slack notification options.
type Notify struct {
	SlackToken   string `yaml:"slack_token"`   // SlackToken is the Bot User OAuth token used to post messages.
	SlackChannel string `yaml:"slack_channel"` // SlackChannel is the channel name or ID where messages will be sent.
	Web          string `yaml:"web"`           // Web is the URL to the web UI.
}
