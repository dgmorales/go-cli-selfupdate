/*
Copyright © 2022 Diego Morales <dgmorales@gmail.com>

*/
package cmd

import (
	"os"

	"github.com/dgmorales/go-cli-selfupdate/kube"
	"github.com/dgmorales/go-cli-selfupdate/version"
	"github.com/spf13/cobra"
)

var flagDebug bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-cli-selfupdate",
	Short: "A test self updatable CLI application",
	Long: `A toy project of a self updatable CLI application, with a K8S twist.

The twist is about interacting with the kubernetes API.
In this test, we will get version information from a ConfigMap
stored in kubernetes.`,
	Version: version.Current,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	kube.Flags.AddFlags(rootCmd.PersistentFlags())
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Activates debug mode. May log very verbose output to stderr")
}
