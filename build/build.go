package build

import (
	"fmt"
	"runtime/debug"
	"time"
)

var (
	// Provided via ldflags (-X)
	Repo    = "https://github.com/ioSpark/consensus"
	Version = "development"

	Commit      string
	CommitShort string
	Date        time.Time
	Modified    bool
)

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, b := range info.Settings {
			switch b.Key {
			case "vcs.revision":
				Commit = b.Value
			case "vcs.modified":
				Modified = b.Value == "true"
			case "vcs.time":
				var err error
				Date, err = time.Parse(time.RFC3339, b.Value)
				if err != nil {
					panic(fmt.Sprintf("Date not in RFC3339 format: %s", b.Value))
				}
			}
		}
	}

	// If we aren't in a VCS, the above won't be set. To prevent errors, set dummy
	// values.
	if Commit == "" {
		Commit = "no-commit"
		CommitShort = "N/A"
	} else {
		CommitShort = Commit[:7]
	}
}
