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
	var stdout, stderr bytes.Buffer

	err := r.runCommand("diff", nil, &stdout, &stderr, true)
	if err != nil && !isAcceptableExitCode(err, 0, 2) {
		return nil, fmt.Errorf("snapraid diff failed: %v\nstderr: %s", err, stderr.String())
	}

	var lines []string
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

// runSync executes `snapraid sync`.
func (r *Runner) runSync() error {
	var stdout, stderr bytes.Buffer

	err := r.runCommand("sync", nil, &stdout, &stderr, false)
	if err != nil && !isAcceptableExitCode(err, 0) {
		return fmt.Errorf("snapraid sync failed: %v\nstderr: %s", err, stderr.String())
	}
	return nil
}

// runTouch executes `snapraid touch`.
func (r *Runner) runTouch() error {
	var stdout, stderr bytes.Buffer

	err := r.runCommand("touch", nil, &stdout, &stderr, false)
	if err != nil && !isAcceptableExitCode(err, 0) {
		return fmt.Errorf("snapraid touch failed: %v\nstderr: %s", err, stderr.String())
	}
	return nil
}

// runScrub executes `snapraid scrub` with --plan and --older-than.
func (r *Runner) runScrub() error {
	var stdout, stderr bytes.Buffer

	args := []string{
		"-plan", strconv.Itoa(r.ScrubPlan),
		"-older-than", strconv.Itoa(r.ScrubOlder),
	}

	err := r.runCommand("scrub", args, &stdout, &stderr, false)
	if err != nil && !isAcceptableExitCode(err, 0) {
		return fmt.Errorf("snapraid scrub failed: %v\nstderr: %s", err, stderr.String())
	}
	return nil
}

// runSmart executes `snapraid smart`.
func (r *Runner) runSmart() error {
	var stdout, stderr bytes.Buffer

	err := r.runCommand("smart", nil, &stdout, &stderr, false)
	if err != nil && !isAcceptableExitCode(err, 0) {
		return fmt.Errorf("snapraid smart failed: %v\nstderr: %s", err, stderr.String())
	}
	return nil
}

// runCommand is the low-level wrapper for invoking snapraid with arguments.
func (r *Runner) runCommand(cmd string, args []string, stdout, stderr io.Writer, quiet bool) error {
	baseArgs := []string{"--conf", r.ConfigFile}
	if quiet {
		baseArgs = append(baseArgs, "--quiet")
	}
	fullArgs := append([]string{cmd}, append(baseArgs, args...)...)

	c := exec.Command(r.SnapraidBin, fullArgs...)
	c.Stdout = stdout
	c.Stderr = stderr

	return c.Run()
}
