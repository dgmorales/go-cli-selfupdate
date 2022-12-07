/*
Copyright © 2022 Diego Morales <dgmorales@gmail.com>

*/
package main

import (
	"log"

	"github.com/dgmorales/go-cli-selfupdate/cmd"
	"github.com/dgmorales/go-cli-selfupdate/logger"
)

func main() {

	fd, err := logger.SetUp()
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer fd.Close()

	cmd.Execute()
}
