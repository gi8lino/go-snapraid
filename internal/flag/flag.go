package flag

import (
	"bytes"
	"fmt"
	"os"

	"github.com/gi8lino/go-snapraid/internal/logging"
	flag "github.com/spf13/pflag"
)

// HelpRequested indicates that the user explicitly requested help or version output.
type HelpRequested struct {
	Message string // Message is the usage or version string to display.
}

func (e *HelpRequested) Error() string { return e.Message }

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
	fs := flag.NewFlagSet("snapraid-runner", flag.ContinueOnError)
	fs.SortFlags = false

	// Basic
	configFile := fs.String("config", "/etc/snapraid-runner.yml", "Path to snapraid runner config")
	verbose := fs.BoolP("verbose", "v", false, "Enable verbose logging")
	dryRun := fs.Bool("dry-run", false, "Skip sync and only perform dry run")
	outputDir := fs.String("output-dir", "", "Directory to write JSON result output")

	// Notifications
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

	// Help/version flags
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
		return Options{}, &HelpRequested{Message: fmt.Sprintf("%s version: %s\n", fs.Name(), version)}
	}

	if showHelp {
		var buf bytes.Buffer
		fs.SetOutput(&buf)
		fs.Usage()
		return Options{}, &HelpRequested{Message: buf.String()}
	}

	// Mutual exclusivity checks for step flags
	if fs.Changed("touch") && fs.Changed("no-touch") {
		return Options{}, fmt.Errorf("cannot use both --touch and --no-touch")
	}
	if fs.Changed("scrub") && fs.Changed("no-scrub") {
		return Options{}, fmt.Errorf("cannot use both --scrub and --no-scrub")
	}
	if fs.Changed("smart") && fs.Changed("no-smart") {
		return Options{}, fmt.Errorf("cannot use both --smart and --no-smart")
	}

	// Resolve step toggles: explicit "no-" flags override enables
	steps := StepsOptions{
		NoTouch: *touch && !*noTouch,
		NoScrub: *scrub && !*noScrub,
		NoSmart: *smart && !*noSmart,
	}

	// Resolve thresholds: enabled by default unless a "no-threshold-<type>" flag was set
	thresholds := ThresholdOptions{
		NoAdd:     !*noAdd,
		NoRemove:  !*noDel,
		NoUpdate:  !*noUp,
		NoCopy:    !*noCp,
		NoMove:    !*noMv,
		NoRestore: !*noRs,
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

// Validate ensures that the specified configuration file exists on disk.
func (o Options) Validate() error {
	if _, err := os.Stat(o.ConfigFile); err != nil {
		return fmt.Errorf("snapraid config file not found: %s", o.ConfigFile)
	}
	return nil
}
