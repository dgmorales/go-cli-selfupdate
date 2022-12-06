/*
Copyright © 2022 Diego Morales <dgmorales@gmail.com>

*/
package version

import (
	_ "embed"
)

//go:embed version.txt
var Version string
