/*
Copyright © 2022 Diego Morales <dgmorales@gmail.com>

*/
package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/dgmorales/go-cli-selfupdate/cmd"
)

func main() {
	logfile := fmt.Sprintf("%s/%s.log", os.TempDir(), path.Base(os.Args[0]))
	f, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		log.Printf("Error opening logfile %s, %v", logfile, err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Printf("Starting %s run", os.Args[0])

	cmd.Execute()
}
