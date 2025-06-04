package snapraid

// Snapraid defines the five low‐level subcommand methods.
type Snapraid interface {
	Touch() error            // Touch runs `snapraid touch`
	Diff() ([]string, error) // Diff runs `snapraid diff` and returns all output lines
	Sync() error             // Sync runs `snapraid sync`
	Scrub() error            // Scrub runs `snapraid scrub` with plan/older‐than flags
	Smart() error            // Smart runs `snapraid smart`
}
