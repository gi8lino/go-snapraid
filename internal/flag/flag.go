package flag

import (
	"fmt"
	"os"

	"github.com/gi8lino/go-snapraid/internal/logging"

	"github.com/containeroo/tinyflags"
)

// ThresholdOptions defines which threshold checks are enabled or disabled.
type ThresholdOptions struct {
	NoAdd     bool // Add controls whether the "added files" threshold check is active.
	NoRemove  bool // Remove controls whether the "removed files" threshold check is active.
	NoUpdate  bool // Update controls whether the "updated files" threshold check is active.
	NoCopy    bool // Copy controls whether the "copied files" threshold check is active.
	NoMove    bool // Move controls whether the "moved files" threshold check is active.
	NoRestore bool // Restore controls whether the "restored files" threshold check is active.
}

// StepsOptions defines which SnapRAID subcommands ("touch", "scrub", "smart") should run.
type StepsOptions struct {
	NoTouch bool // Touch enables the "snapraid touch" step.
	NoScrub bool // Scrub enables the "snapraid scrub" step.
	NoSmart bool // Smart enables the "snapraid smart" step.
}

// Options holds all configuration values parsed from CLI flags.
type Options struct {
	LogFormat  logging.LogFormat // LogFormat determines the output format (e.g. text or JSON) for logging.
	ConfigFile string            // ConfigFile is the path to the YAML configuration file for snapraid-runner.
	DryRun     bool              // DryRun, when true, skips actual sync/scrub/smart steps (simulates only).
	Verbose    bool              // Verbose, when true, enables verbose logging output.
	OutputDir  string            // OutputDir is the directory where JSON result files will be written (if non-empty).
	NoNotify   bool              // NoNotify disables Slack notifications when true.
	Steps      StepsOptions      // Steps contains which SnapRAID subcommands ("touch", "scrub", "smart") to execute.
	Thresholds ThresholdOptions  // Thresholds contains which threshold checks (add/remove/update/…) are enabled.
	ScrubPlan  int               // ScrubPlan is the percentage (0–100) passed to the "scrub" subcommand.
	ScrubOlder int               // ScrubOlder is the "older-than" age (in days) passed to the "scrub" subcommand.
}

// ParseFlags parses CLI flags into a structured Options instance. It also handles
// --help and --version, returning a HelpRequested error when appropriate.
func ParseFlags(args []string, version string) (Options, error) {
	opts := Options{}
	tf := tinyflags.NewFlagSet("snapraid-runner", tinyflags.ContinueOnError)
	tf.Version(version)

	// Basic
	tf.StringVar(&opts.ConfigFile, "config", "/etc/snapraid-runner.yml", "Path to snapraid runner config").
		Value()
	tf.BoolVar(&opts.Verbose, "verbose", false, "Enable verbose logging").
		Short("v").
		Value()
	tf.BoolVar(&opts.DryRun, "dry-run", false, "Skip sync and only perform dry run").Value()
	tf.StringVar(&opts.OutputDir, "output-dir", "", "Directory to write JSON result output").Value()
	logFormat := tf.String("log-format", "text", "Log format").
		Choices("text", "json").
		HideAllowed().
		Short("l").
		Value()

	// Notifications
	tf.BoolVar(&opts.NoNotify, "no-notify", false, "Disable Slack notifications").Value()

	// Step toggles
	touch := tf.Bool("touch", false, "Enable touch step").
		OneOfGroup("steps").
		Value()
	noTouch := tf.Bool("no-touch", false, "Disable touch step").
		OneOfGroup("steps").
		Value()

	scrub := tf.Bool("scrub", false, "Enable scrub step").
		OneOfGroup("scrub").
		Value()
	noScrub := tf.Bool("no-scrub", false, "Disable scrub step").
		OneOfGroup("scrub").
		Value()

	smart := tf.Bool("smart", false, "Enable smart step").
		OneOfGroup("smart").
		Value()
	noSmart := tf.Bool("no-smart", false, "Disable smart step").
		OneOfGroup("smart").
		Value()

	// Threshold disablers
	noAdd := tf.Bool("no-threshold-add", false, "Disable threshold check for added files").Value()
	noDel := tf.Bool("no-threshold-del", false, "Disable threshold check for removed files").Value()
	noUp := tf.Bool("no-threshold-up", false, "Disable threshold check for updated files").Value()
	noCp := tf.Bool("no-threshold-cp", false, "Disable threshold check for copied files").Value()
	noMv := tf.Bool("no-threshold-mv", false, "Disable threshold check for moved files").Value()
	noRs := tf.Bool("no-threshold-rs", false, "Disable threshold check for restored files").Value()

	// Scrub options
	tf.IntVar(&opts.ScrubPlan, "plan", 22, "Scrub plan percentage (0–100)").
		Validate(func(i int) error {
			if i < 0 || i > 100 {
				return fmt.Errorf("scrub plan must be between 0 and 100")
			}
			return nil
		}).
		Value()
	tf.IntVar(&opts.ScrubOlder, "older-than", 12, "Scrub files older than N days").Value()

	// Parse args
	if err := tf.Parse(args); err != nil {
		return Options{}, err
	}

	// Resolve step toggles: explicit "no-" flags override enables
	opts.Steps = StepsOptions{
		NoTouch: *touch && !*noTouch,
		NoScrub: *scrub && !*noScrub,
		NoSmart: *smart && !*noSmart,
	}

	// Resolve log format
	opts.LogFormat = logging.LogFormat(*logFormat)

	// Resolve thresholds: enabled by default unless a "no-threshold-<type>" flag was set
	opts.Thresholds = ThresholdOptions{
		NoAdd:     !*noAdd,
		NoRemove:  !*noDel,
		NoUpdate:  !*noUp,
		NoCopy:    !*noCp,
		NoMove:    !*noMv,
		NoRestore: !*noRs,
	}

	return opts, nil
}

// Validate ensures that the specified configuration file exists on disk.
func (o Options) Validate() error {
	if _, err := os.Stat(o.ConfigFile); err != nil {
		return fmt.Errorf("snapraid config file not found: %s", o.ConfigFile)
	}
	return nil
}
