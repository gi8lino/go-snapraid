# SnapRAID Runner

SnapRAID Runner is a command-line utility that automates SnapRAID operations, including syncing, scrubbing, touch, and smart checks. It wraps the SnapRAID binary with additional threshold checks, configurable steps, and optional Slack notifications. Configuration is provided via a YAML file, and all behaviors can be adjusted through CLI flags.

## Features

- **Configurable Steps**: Enable or disable SnapRAID subcommands (`touch`, `scrub`, `smart`) individually.
- **Threshold Checks**: Prevent sync if file changes exceed configured thresholds (added, removed, updated, copied, moved, restored).
- **Dry-Run Mode**: Perform a dry run to preview actions without performing any sync.
- **Verbose Logging**: Toggle detailed logging output.
- **Output Directory**: Write JSON-formatted results to a directory for further processing.
- **Slack Notifications**: Send notifications to a Slack channel after completion (configurable via the YAML file).
- **Configuration Validation**: Ensures required binaries and configuration files exist before execution.

## Configuration

SnapRAID Runner reads a YAML configuration file to determine command behaviors, thresholds, and notification settings. By default, it looks for `/etc/snapraid_runner.conf`, but you can override this via the `--conf` flag.

Below is an example configuration file (`snapraid_runner.conf`):

```yaml
# Path to the SnapRAID binary
snapraid_bin: /usr/bin/snapraid

# Path to the SnapRAID configuration file (.conf)
snapraid_config: /etc/snapraid.conf

# Directory where JSON results will be written
output_dir: /var/log/go-snapraid/

# Threshold limits before blocking SnapRAID sync
thresholds:
  add: 100 # Maximum number of added files
  remove: 50 # Maximum number of removed files
  update: 200 # Maximum number of updated files
  copy: 150 # Maximum number of copied files
  move: 75 # Maximum number of moved files
  restore: 25 # Maximum number of restored files

# Steps to run: set to true or false
steps:
  touch: true # Enable `snapraid touch`
  scrub: true # Enable `snapraid scrub`
  smart: true # Enable `snapraid smart`

# Scrub options (only used if 'scrub: true')
scrub:
  plan: 22 # Scrub plan percentage (0–100)
  older_than: 12 # Scrub files older than N days

# Slack notification settings
notifications:
  slack_token: xoxb-1234567890-abcdefg # Slack Bot User OAuth Token
  slack_channel: "#snapraid" # Channel name or ID
```

- **`snapraid_bin`**: Full path to the `snapraid` executable.
- **`snapraid_config`**: Path to the SnapRAID config file used by the `snapraid` command.
- **`output_dir`**: Directory for writing JSON result files. If unset, JSON output is not written.
- **`thresholds`**: Numeric limits for each file-change category. If any threshold is exceeded, SnapRAID sync is aborted.
- **`steps.touch`**, **`steps.scrub`**, **`steps.smart`**: Boolean flags determining which SnapRAID subcommands run.
- **`scrub.plan`**, **`scrub.older_than`**: Parameters for the `snapraid scrub` command, used only if `steps.scrub` is true.
- **`notifications.slack_token`**, **`notifications.slack_channel`**: Credentials and channel for sending a Slack notification after execution. If `slack_token` or `slack_channel` is empty, notifications are disabled.

## Usage

```bash
go-snapraid [flags]
```

### Common Flags

```
--config <path>         Path to YAML configuration file (default: /etc/snapraid_runner.conf)
--verbose, -v           Enable verbose logging
--dry-run               Skip sync; only perform a dry-run check
--no-notify             Disable Slack notifications
--output-dir <dir>      Directory to write JSON-formatted result output

--touch                 Enable the `touch` step
--no-touch              Disable the `touch` step
--scrub                 Enable the `scrub` step
--no-scrub              Disable the `scrub` step
--smart                 Enable the `smart` step
--no-smart              Disable the `smart` step

--no-threshold-add      Disable threshold check for added files
--no-threshold-del      Disable threshold check for removed files
--no-threshold-up       Disable threshold check for updated files
--no-threshold-cp       Disable threshold check for copied files
--no-threshold-mv       Disable threshold check for moved files
--no-threshold-rs       Disable threshold check for restored files

--plan <int>            Scrub plan percentage (0–100) (default: 22)
--older-than <int>      Scrub files older than N days (default: 12)

--help, -h              Show help and exit
--version               Show version and exit
```

- If both an enabling flag (e.g., `--scrub`) and its disabling counterpart (e.g., `--no-scrub`) are provided, the program exits with an error.
- Threshold checks are enabled by default; use `--no-threshold-*` flags to disable specific checks.
- To see usage and flag descriptions, run:

  ```bash
  go-snapraid --help
  ```

### Examples

1. **Run with default configuration**
   Uses `/etc/snapraid_runner.conf` and runs all steps specified in the YAML:

   ```bash
   go-snapraid
   ```

2. **Dry run, verbose mode, custom config**
   Override default config, enable verbose output, and perform a dry-run:

   ```bash
   go-snapraid --config /home/user/.go-snapraid.yml --verbose --dry-run
   ```

3. **Disable scrub step and disable threshold checks for added files**
   Run SnapRAID without scrubbing and skip added-file threshold:

   ```bash
   go-snapraid --no-scrub --no-threshold-add
   ```

4. **Write JSON output to a specific directory**
   Specify an output directory to store JSON results:

   ```bash
   go-snapraid --output-dir /tmp/go-snapraid
   ```

5. **Show version and exit**

   ```bash
   go-snapraid --version
   ```

### Configuration File Location

By default, SnapRAID Runner looks for its configuration at:

```text
/etc/go-snapraid.yml
```

To override the location, use the `--config` flag:

```bash
go-snapraid --conf /path/to/custom_config.yml
```

If the specified config file does not exist, the program exits with an error:

```bash
go-snapraid: snapraid config file not found: /path/to/custom_config.yml
```

### JSON Output

If the `--output-dir` flag is specified (or `output_dir` in the YAML is set), SnapRAID Runner will write a JSON file containing:

- **Timestamp**: Time of execution
- **Executed Steps**: Which subcommands ran (`touch`, `scrub`, `smart`)
- **Threshold Results**: Counts for added, removed, updated, copied, moved, and restored files, and whether thresholds passed or failed
- **SnapRAID Exit Codes**: Exit codes for each SnapRAID command executed
- **Errors or Warnings**: Any errors or warnings encountered during execution

These files will be named using the UTC timestamp, for example:

```
2025-06-02T14-30-00Z_result.json
```

### Slack Notifications

If Slack notifications are configured in the YAML file (non-empty `slack_token` and `slack_channel`), SnapRAID Runner will send a JSON payload to Slack summarizing:

- Execution status (success or failure)
- Threshold check results
- SnapRAID exit statuses
- Any errors encountered

To temporarily disable Slack notifications, pass `--no-notify` on the command line.

## License

SnapRAID Runner is licensed under the MIT License. See [LICENSE](./LICENSE) for details.
