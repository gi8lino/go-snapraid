package snapraid

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
)

// runDiff executes `snapraid diff` and returns the stdout+stderr lines.
// It logs each line immediately (via r.Logger), and also keeps a copy in `buf`.
func (r *Runner) runDiff() ([]string, error) {
	// Create a temporary buffer to capture the full output of this one command:
	var buf bytes.Buffer

	// Create a loggerWriter that tags each line under "diff"
	lw := newLoggerWriter(r.Logger, "diff")

	// Build a MultiWriter that duplicates everything into both:
	//      - buf       (for post‐scan / filtering)
	//      - lw        (which immediately logs each newline‐terminated line)
	combined := io.MultiWriter(&buf, lw)

	// Run the command, sending both stdout & stderr into `combined`:
	err := r.runCommand("diff", nil, combined, true)
	if err != nil && !isAcceptableExitCode(err, 0, 2) {
		// In case of an unacceptable exit code, we still have the full contents of buf.
		// We can log them one more time at ERROR level, or just return them in the error.
		return nil, fmt.Errorf("snapraid diff failed: %v\noutput: %s", err, buf.String())
	}

	// Now that the process has exited, `buf` contains all interleaved output.
	// We scan `buf` to build []string if the caller wants to process or return it.
	var lines []string
	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

// runSync executes `snapraid sync`.
func (r *Runner) runSync() error {
	var buf bytes.Buffer
	lw := newLoggerWriter(r.Logger, "sync")
	combined := io.MultiWriter(&buf, lw)

	err := r.runCommand("sync", nil, combined, false)
	if err != nil && !isAcceptableExitCode(err, 0) {
		return fmt.Errorf("snapraid sync failed: %v\noutput: %s", err, buf.String())
	}
	// We don’t need to scan buf if we don’t have any []string to return,
	// but we keep `buf` around in case we want to include it in the error above.
	return nil
}

// runTouch executes `snapraid touch`.
func (r *Runner) runTouch() error {
	var buf bytes.Buffer
	lw := newLoggerWriter(r.Logger, "touch")
	combined := io.MultiWriter(&buf, lw)

	err := r.runCommand("touch", nil, combined, false)
	if err != nil && !isAcceptableExitCode(err, 0) {
		return fmt.Errorf("snapraid touch failed: %v\noutput: %s", err, buf.String())
	}
	return nil
}

// runScrub executes `snapraid scrub` with --plan and --older-than.
func (r *Runner) runScrub() error {
	var buf bytes.Buffer
	lw := newLoggerWriter(r.Logger, "scrub")
	combined := io.MultiWriter(&buf, lw)

	args := []string{
		"-plan", strconv.Itoa(r.ScrubPlan),
		"-older-than", strconv.Itoa(r.ScrubOlder),
	}
	err := r.runCommand("scrub", args, combined, false)
	if err != nil && !isAcceptableExitCode(err, 0) {
		return fmt.Errorf("snapraid scrub failed: %v\noutput: %s", err, buf.String())
	}
	return nil
}

// runSmart executes `snapraid smart`.
func (r *Runner) runSmart() error {
	var buf bytes.Buffer
	lw := newLoggerWriter(r.Logger, "smart")
	combined := io.MultiWriter(&buf, lw)

	err := r.runCommand("smart", nil, combined, false)
	if err != nil && !isAcceptableExitCode(err, 0) {
		return fmt.Errorf("snapraid smart failed: %v\noutput: %s", err, buf.String())
	}
	return nil
}

// runCommand is the low-level wrapper for invoking snapraid with arguments.
func (r *Runner) runCommand(cmd string, args []string, output io.Writer, quiet bool) error {
	baseArgs := []string{"--conf", r.ConfigFile}
	if quiet {
		baseArgs = append(baseArgs, "--quiet")
	}
	fullArgs := append([]string{cmd}, append(baseArgs, args...)...)

	c := exec.Command(r.SnapraidBin, fullArgs...)
	c.Stdout = output
	c.Stderr = output

	return c.Run()
}
