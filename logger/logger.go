package logger

import (
	"fmt"
	"log"
	"os"
	"path"
)

var WorkDir = path.Join(os.TempDir(), "go-cli-selfupdate")
var LogFile string

// SetUp sets up error logging by creating logfiles, configuring log lib, etc.
func SetUp() (*os.File, error) {
	err := os.MkdirAll(WorkDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating logdir %s: %w", WorkDir, err)
	}

	LogFile = fmt.Sprintf("%s/%s.log", WorkDir, path.Base(os.Args[0]))
	fd, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening logfile %s: %w", LogFile, err)
	}

	log.SetOutput(fd)
	log.Printf("Starting %s run", os.Args[0])

	return fd, nil
}
