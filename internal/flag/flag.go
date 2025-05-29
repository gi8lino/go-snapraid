package flag

import (
	"bytes"
	"fmt"
	"os"

	"github.com/gi8lino/go-snapraid/internal/logging"
	flag "github.com/spf13/pflag"
)

// HelpRequested indicates help was explicitly requested.
type HelpRequested struct {
	Message string
}

func (e *HelpRequested) Error() string { return e.Message }

// ThresholdOptions defines which threshold checks are enabled.
type ThresholdOptions struct {
	Add     bool
	Remove  bool
	Update  bool
	Copy    bool
	Move    bool
	Restore bool
}

// StepsOptions defines which snapraid steps to run.
type StepsOptions struct {
	Touch bool
	Scrub bool
	Smart bool
}

// Options holds all configuration values parsed from CLI flags.
type Options struct {
	LogFormat  logging.LogFormat
	ConfigFile string
	DryRun     bool
	Verbose    bool
	OutputDir  string
	NoNotify   bool

	Steps      StepsOptions
	Thresholds ThresholdOptions
	ScrubPlan  int
	ScrubOlder int
}

// ParseFlags parses CLI flags into structured Options.
func ParseFlags(args []string, version string) (Options, error) {
	fs := flag.NewFlagSet("snapraid-runner", flag.ContinueOnError)
	fs.SortFlags = false

	// Basic
	configFile := fs.String("conf", "/etc/snapraid_runner.conf", "Path to snapraid config")
	verbose := fs.BoolP("verbose", "v", false, "Enable verbose logging")
	dryRun := fs.Bool("dry-run", false, "Skip sync and only perform dry run")
	noNotify := fs.Bool("no-notify", false, "Disable Slack notifications")

	// Step toggles
	touch := fs.Bool("touch", false, "Enable touch step")
	noTouch := fs.Bool("no-touch", false, "Disable touch step")
	scrub := fs.Bool("scrub", false, "Enable scrub step")
	noScrub := fs.Bool("no-scrub", false, "Disable scrub step")
	smart := fs.Bool("smart", false, "Enable smart step")
	noSmart := fs.Bool("no-smart", false, "Disable smart step")

	// Threshold disablers
	noAdd := fs.Bool("no-threshold-add", false, "Disable threshold check for added files")
	noDel := fs.Bool("no-threshold-del", false, "Disable threshold check for removed files")
	noUp := fs.Bool("no-threshold-up", false, "Disable threshold check for updated files")
	noCp := fs.Bool("no-threshold-cp", false, "Disable threshold check for copied files")
	noMv := fs.Bool("no-threshold-mv", false, "Disable threshold check for moved files")
	noRs := fs.Bool("no-threshold-rs", false, "Disable threshold check for restored files")

	// Scrub options
	scrubPlan := fs.Int("plan", 22, "Scrub plan percentage (0–100)")
	scrubOlder := fs.Int("older-than", 12, "Scrub files older than N days")

	// Notifications
	outputDir := fs.String("output-dir", "", "Directory to write JSON result output")

	// Help/version
	var showHelp, showVersion bool
	fs.BoolVarP(&showHelp, "help", "h", false, "Show help and exit")
	fs.BoolVar(&showVersion, "version", false, "Show version and exit")

	// Custom usage output
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [flags]\n\nFlags:\n", fs.Name()) // nolint:errcheck
		fs.PrintDefaults()
	}

	// Parse args
	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}

	if showVersion {
		return Options{}, &HelpRequested{Message: fmt.Sprintf("%s version %s\n", fs.Name(), version)}
	}

	if showHelp {
		var buf bytes.Buffer
		fs.SetOutput(&buf)
		fs.Usage()
		return Options{}, &HelpRequested{Message: buf.String()}
	}

	// Mutual exclusivity checks
	if fs.Changed("touch") && fs.Changed("no-touch") {
		return Options{}, fmt.Errorf("cannot use both --touch and --no-touch")
	}
	if fs.Changed("scrub") && fs.Changed("no-scrub") {
		return Options{}, fmt.Errorf("cannot use both --scrub and --no-scrub")
	}
	if fs.Changed("smart") && fs.Changed("no-smart") {
		return Options{}, fmt.Errorf("cannot use both --smart and --no-smart")
	}

	// Step resolution
	steps := StepsOptions{
		Touch: *touch && !*noTouch,
		Scrub: *scrub && !*noScrub,
		Smart: *smart && !*noSmart,
	}

	// Threshold resolution (enabled unless explicitly disabled)
	thresholds := ThresholdOptions{
		Add:     !*noAdd,
		Remove:  !*noDel,
		Update:  !*noUp,
		Copy:    !*noCp,
		Move:    !*noMv,
		Restore: !*noRs,
	}

	return Options{
		LogFormat:  logging.LogFormatText,
		ConfigFile: *configFile,
		DryRun:     *dryRun,
		Verbose:    *verbose,
		OutputDir:  *outputDir,
		NoNotify:   *noNotify,
		Steps:      steps,
		Thresholds: thresholds,
		ScrubPlan:  *scrubPlan,
		ScrubOlder: *scrubOlder,
	}, nil
}

// Validate ensures required binaries and config files exist.
func (o Options) Validate() error {
	if _, err := os.Stat(o.ConfigFile); err != nil {
		return fmt.Errorf("snapraid config file not found: %s", o.ConfigFile)
	}
	return nil
}
