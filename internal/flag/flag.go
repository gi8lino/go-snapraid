package flag

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/gi8lino/go-snapraid/internal/logging"
	flag "github.com/spf13/pflag"
)

// HelpRequested indicates help was explicitly requested.
type HelpRequested struct {
	Message string
}

func (e *HelpRequested) Error() string { return e.Message }

// Options holds CLI configuration values.
type Options struct {
	LogFormat    logging.LogFormat
	SnapraidBin  string
	ConfigFile   string
	DryRun       bool
	Verbose      bool
	ThresholdAdd int
	ThresholdDel int
	ThresholdUp  int
	ThresholdCp  int
	ThresholdMv  int
	ThresholdRs  int
	OutputDir    string
	SlackToken   string
	SlackChan    string
}

// ParseFlags parses CLI flags.
func ParseFlags(args []string, version string) (Options, error) {
	fs := flag.NewFlagSet("snapraid-runner", flag.ContinueOnError)
	fs.SortFlags = false

	snapraidBin := fs.String("bin", "/usr/bin/snapraid", "Path to snapraid executable")
	config := fs.String("conf", "/etc/snapraid.conf", "Path to snapraid config")
	verbose := fs.BoolP("verbose", "v", false, "Enable verbose output")
	dryRun := fs.Bool("dry-run", false, "Dry run, don't run sync")

	// Thresholds
	thresholdAdd := fs.Int("add-threshold", 100, "Max allowed added files before abort")
	thresholdDel := fs.Int("remove-threshold", 50, "Max allowed removed files before abort")
	thresholdUp := fs.Int("update-threshold", 50, "Max allowed updated files before abort")
	thresholdCp := fs.Int("copy-threshold", 50, "Max allowed copied files before abort")
	thresholdMv := fs.Int("move-threshold", 50, "Max allowed moved files before abort")
	thresholdRs := fs.Int("restore-threshold", 50, "Max allowed restored files before abort")

	// Notifications
	outputDir := fs.String("output-dir", "/var/log/snapraid-runner", "Directory to write JSON run results")
	slackToken := fs.String("slack-token", "", "Slack bot token")
	slackChan := fs.String("slack-channel", "", "Slack channel ID")

	var showHelp, showVersion bool
	fs.BoolVarP(&showHelp, "help", "h", false, "Show help and exit")
	fs.BoolVar(&showVersion, "version", false, "Print version and exit")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [flags]\n\nFlags:\n", strings.ToLower(fs.Name())) // nolint:errcheck
		fs.PrintDefaults()
	}

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

	return Options{
		SnapraidBin:  *snapraidBin,
		ConfigFile:   *config,
		Verbose:      *verbose,
		DryRun:       *dryRun,
		ThresholdAdd: *thresholdAdd,
		ThresholdDel: *thresholdDel,
		ThresholdUp:  *thresholdUp,
		ThresholdCp:  *thresholdCp,
		ThresholdMv:  *thresholdMv,
		ThresholdRs:  *thresholdRs,
		OutputDir:    *outputDir,
		SlackToken:   *slackToken,
		SlackChan:    *slackChan,
	}, nil
}

// Validate checks the configuration for errors.
func (o Options) Validate() error {
	if _, err := os.Stat(o.SnapraidBin); err != nil {
		return fmt.Errorf("snapraid executable not found at %s", o.SnapraidBin)
	}
	if _, err := os.Stat(o.ConfigFile); err != nil {
		return fmt.Errorf("snapraid config not found at %s", o.ConfigFile)
	}
	return nil
}
