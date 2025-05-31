package snapraid

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
)

// runDiff executes `snapraid diff` and returns the stdout lines.
func (r *Runner) runDiff() ([]string, error) {
	// Create a temporary buffer to capture the full output of this one command:
	var buf bytes.Buffer

	// Set up a MultiWriter so that:
	//    - everything still goes into 'buf' (so we can scan/filter after)
	//    - and also goes into r.Output (which itself might be os.Stdout or another sink)
	combined := io.MultiWriter(&buf, r.Output)

	// Run the command, sending both stdout+stderr into 'combined':
	err := r.runCommand("diff", nil, combined, true)
	if err != nil && !isAcceptableExitCode(err, 0, 2) {
		return nil, fmt.Errorf("snapraid diff failed: %v\nstderr: %s", err, buf.String())
	}

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
	combined := io.MultiWriter(&buf, r.Output)

	err := r.runCommand("sync", nil, combined, false)
	if err != nil && !isAcceptableExitCode(err, 0) {
		return fmt.Errorf("snapraid sync failed: %v\noutput: %s", err, buf.String())
	}
	return nil
}

// runTouch executes `snapraid touch`.
func (r *Runner) runTouch() error {
	var buf bytes.Buffer
	combined := io.MultiWriter(&buf, r.Output)

	err := r.runCommand("touch", nil, combined, false)
	if err != nil && !isAcceptableExitCode(err, 0) {
		return fmt.Errorf("snapraid touch failed: %v\noutput: %s", err, buf.String())
	}
	return nil
}

// runScrub executes `snapraid scrub` with --plan and --older-than.
func (r *Runner) runScrub() error {
	var buf bytes.Buffer
	combined := io.MultiWriter(&buf, r.Output)

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
	combined := io.MultiWriter(&buf, r.Output)

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
