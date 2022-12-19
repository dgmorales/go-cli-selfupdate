/*
Copyright © 2022 Diego Morales <dgmorales@gmail.com>

*/
package version

import (
	_ "embed"
)

//go:embed version.txt
var Current string

// Assertion informs if our version is latest, can or must update.
//
// Its values may be used as shell exit status codes for a version check command.
type Assertion int

// won't use iota bellow because values matter for use as exit codes
// Avoid using 1 and 2 that are frequently used as more general error exit codes for
// commands: https://tldp.org/LDP/abs/html/exitcodes.html
const (
	IsLatest   Assertion = 0
	CanUpdate  Assertion = 10
	MustUpdate Assertion = 20
	IsBeyond   Assertion = 30  // Beyond current version: development version
	IsUnknown  Assertion = 60
)

type Checker interface {
	Minimal() string
	Current() string
	Latest() string
	Check() (Assertion)
	DownloadLatest() (filename string, err error)
}
