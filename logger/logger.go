package logger

import (
	"io/ioutil"
	"log"
	"os"
	"path"
)

var WorkDir = path.Join(os.TempDir(), "go-cli-selfupdate")
var LogFile string

// SetUp sets up logging
//
// Our logging for now is extremely simple:
// it logs to stderr when in debug mode, otherwise discard log message
func SetUp(debug bool) {
	if debug {
		log.SetOutput(os.Stderr)
		log.Printf("%s is in debug mode.", os.Args[0])
	} else {
		log.SetOutput(ioutil.Discard)
	}
}
