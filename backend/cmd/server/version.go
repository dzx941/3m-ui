package main

import "fmt"

var (
	version   = "v0.1.0-rc.1"
	gitCommit = "unknown"
	buildTime = "unknown"
)

func versionString() string {
	return fmt.Sprintf("3m-ui %s\ngit commit: %s\nbuild time: %s", version, gitCommit, buildTime)
}
