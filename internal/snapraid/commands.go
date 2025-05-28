package snapraid

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"

	"github.com/gi8lino/go-snapraid/internal/flag"
)

// runDiff executes 'snapraid diff' and returns its stdout output line by line.
func runDiff(o flag.Options) ([]string, error) {
	var stdout, stderr bytes.Buffer

	err := runCommand(o, "diff", &stdout, &stderr, true)
	// exit code 0 = no changes, 2 = changes → both are acceptable
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

// runSync executes 'snapraid sync' and returns any non-zero error.
func runSync(o flag.Options) error {
	var stdout, stderr bytes.Buffer

	err := runCommand(o, "sync", &stdout, &stderr, false)
	if err != nil && !isAcceptableExitCode(err, 0) {
		return fmt.Errorf("snapraid sync failed: %v\nstderr: %s", err, stderr.String())
	}
	return nil
}

// runCommand executes a SnapRAID command with the given arguments and I/O streams.
func runCommand(o flag.Options, cmd string, stdout, stderr io.Writer, quiet bool) error {
	args := []string{"--conf", o.ConfigFile}
	if quiet {
		args = append(args, "--quiet")
	}
	args = append([]string{cmd}, args...)

	c := exec.Command(o.SnapraidBin, args...)
	c.Stdout = stdout
	c.Stderr = stderr

	return c.Run()
}
