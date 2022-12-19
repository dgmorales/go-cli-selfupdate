/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dgmorales/go-cli-selfupdate/logger"
	"github.com/dgmorales/go-cli-selfupdate/selfupdate"
	"github.com/dgmorales/go-cli-selfupdate/start"
	"github.com/dgmorales/go-cli-selfupdate/version"
	"github.com/spf13/cobra"
)

var flagCheck bool
var flagYes bool

// selfUpdateCmd represents the selfUpdate command
var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Self update this CLI to latest version",
	Long: `self-update downloads the latest version of this CLI and applies it, overwriting the
current binary.

You may not need to call this directly. The CLI always checks for the newest version
and warns you if there's an update available. If the current version is too old to
talk to our API server, it will force you to update before continuing. It always asks
first, though (unless you pass --yes).

In any case, the CLI will exit after the apply and you will need to call it again to
load the binary with the new version.

You may run this command with --check for just checking if an update is available or
required. Exit code will reflect the update decision:

IsLatest   = 0
CanUpdate  = 1
MustUpdate = 2
`,
	Run: func(cmd *cobra.Command, args []string) {
		state, err := start.ForAPIUse()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		ans := versionCheck(state.Version)

		// versionCheck is meant to run from any Command
		// The bellow switch looks again to the version check assertion and performs
		// user interactions we want only on this command.
		switch ans {
		case version.IsLatest:
			fmt.Printf("You are at the latest version (%s)\n", state.Version.Latest())
		case version.CanUpdate:
			if !flagCheck {
				confirmAndUpdate(ans, state.Version)
			}
		}

		os.Exit(int(ans))
	},
}

func init() {
	rootCmd.AddCommand(selfUpdateCmd)

	selfUpdateCmd.Flags().BoolVarP(&flagCheck, "check", "c", false, "Just check and report if there is a new version available.")
	selfUpdateCmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Don't ask before changing your system. Assume yes.")
}

// Check checks if current version can or must be update, and interacts with the user
// about it
func versionCheck(v version.Checker) version.Assertion {
	decision, err := v.Check()
	// We will just log errors below and continue, without disturbing user interaction
	// flow. Version check and update is a non essential feature.
	if err != nil {
		fmt.Printf("Info: we are having some trouble checking for a new version of the CLI. Check details on logfile %s\n", logger.LogFile)
		log.Println(err)
	}
	log.Printf("debug: cli information dump: %v\n", v)

	switch decision {
	case version.MustUpdate:
		fmt.Printf("Warning: your current version (%s) is not supported anymore (minimal: %s, latest: %s). You need to update it.\n",
			v.Current(), v.Minimal(), v.Latest())

		if !flagCheck {
			confirmAndUpdate(decision, v)
		}

	case version.CanUpdate:
		fmt.Printf("Warning: there's a newer version (%s), but this version (%s) is still usable. You can update it by running %s self-update.\n",
			v.Latest(), v.Current(), os.Args[0])
		// UX decision: just warns, and do not ask the user for update if it is not required.
	}

	return decision
}

// confirmAndUpdate will confirms if we can proceed with the self-update,
// and performs the update if confirmed.
//
// It will ask the user, check for --yes flags, etc.
//
// If the update is **not required** and not performed this function returns.
// Otherwise this function assures the program is terminated.
func confirmAndUpdate(d version.Assertion, v version.Checker) {
	if !flagYes && !askIfUpdate() {
		if d == version.MustUpdate {
			fmt.Println("Cannot continue without updating. Exiting.")
			os.Exit(int(d))
		} else {
			// Update is optional, let the program continue
			return
		}
	}

	fmt.Println("Downloading and applying latest release ...")
	filename, err := v.DownloadLatest()
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		os.Exit(1)
	}
	err = selfupdate.Apply(filename)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		os.Exit(1)
	}

	// Terminate to force the user to load the updated binary
	os.Exit(0)
}

// askIfUpdate will ask the user if we should update now
func askIfUpdate() bool {
	ans := false
	prompt := &survey.Confirm{
		Message: "Can I download the latest version and self-update now?",
		Default: true,
	}
	survey.AskOne(prompt, &ans)

	return ans
}
