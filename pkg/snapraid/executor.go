package snapraid

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strconv"
)

// DefaultExecutor is the real implementation of Snapraid that shells out.
type DefaultExecutor struct {
	configPath string       // path to YAML config (used by "--conf")
	binaryPath string       // path to the snapraid executable
	scrubPlan  int          // percentage (0–100) passed to "scrub"
	scrubOlder int          // days passed to "scrub --older-than"
	logger     *slog.Logger // structured logger for per‐line output
}

// Touch shells out to `snapraid touch` and logs each line under "touch".
func (d *DefaultExecutor) Touch() error {
	return d.runCommand("touch", nil, "touch")
}

// Diff shells out to `snapraid diff`, logs under "diff", and returns all stdout lines.
func (d *DefaultExecutor) Diff() ([]string, error) {
	var buf bytes.Buffer
	lw := newLoggerWriter(d.logger, "diff")
	combined := io.MultiWriter(&buf, lw)

	err := d.runCommandToWriter("diff", nil, combined)
	if err != nil && !isAcceptableExitCode(err, 0, 2) {
		return nil, fmt.Errorf("snapraid diff failed: %v: %s", err, buf.String())
	}

	var lines []string
	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

// Sync shells out to `snapraid sync` and logs each line under "sync".
func (d *DefaultExecutor) Sync() error {
	return d.runCommand("sync", nil, "sync")
}

// Scrub shells out to `snapraid scrub --plan X --older-than Y` under "scrub".
func (d *DefaultExecutor) Scrub() error {
	args := []string{
		"--plan", strconv.Itoa(d.scrubPlan),
		"--older-than", strconv.Itoa(d.scrubOlder),
	}
	return d.runCommand("scrub", args, "scrub")
}

// Smart shells out to `snapraid smart` and logs each line under "smart".
func (d *DefaultExecutor) Smart() error {
	return d.runCommand("smart", nil, "smart")
}

// runCommand runs `snapraid <cmd> [--plan X] [--older-than Y]`, logging under the given tag.
func (d *DefaultExecutor) runCommand(cmd string, args []string, tag string) error {
 var outBuf, errBuf bytes.Buffer
 stdoutLog := newLoggerWriter(d.logger, tag) // could add ".stdout" tag
 stderrLog := newLoggerWriter(d.logger, tag) // or different severity

 stdoutCombined := io.MultiWriter(&outBuf, stdoutLog)
 stderrCombined := io.MultiWriter(&errBuf, stderrLog)

 if err := d.runCommandToWriter(cmd, args, stdoutCombined, stderrCombined); err != nil {
	 return fmt.Errorf("snapraid %s failed: %w\nstderr:\n%s", cmd, err, errBuf.String())
 }
}

// runCommandToWriter builds and invokes the actual `snapraid <cmd> …`, writing stdout+stderr to w.
func (d *DefaultExecutor) runCommandToWriter(cmd string, args []string, stdout, stderr io.Writer) error {
	baseArgs := []string{"--conf", d.configPath, "--quiet"}
	fullArgs := append([]string{cmd}, append(baseArgs, args...)...)

	fmt.Fprintf(stdout, "Running %s\n", cmd)
	c := exec.Command(d.binaryPath, fullArgs...)
	c.Stdout = stdout
	c.Stderr = stderr
	return c.Run()
}
