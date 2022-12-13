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
	"github.com/dgmorales/go-cli-selfupdate/version"
	"github.com/google/go-github/v48/github"
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
		decision := versionCheck()
		if flagCheck {
			os.Exit(int(decision))
		}

		if decision == selfupdate.CanUpdate {
			// TODO: needs to fix confirmAndUpdate signature
			fmt.Printf("will ask for update\n")
		}

		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(selfUpdateCmd)

	selfUpdateCmd.Flags().BoolVarP(&flagCheck, "check", "c", false, "Just check and report if there is a new version available.")
	selfUpdateCmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Don't ask before changing your system. Assume yes.")
}

// Check checks if current version can or must be update, and interacts with the user
// about it
func versionCheck() selfupdate.UpdateDecision {
	i := selfupdate.CliInfo{}
	gh := github.NewClient(nil)
	decision, err := selfupdate.Check(&i, gh)
	// We will just log errors below and continue, without disturbing user interaction
	// flow. Version check and update is a non essential feature.
	if err != nil {
		fmt.Printf("Info: we are having some trouble checking for a new version of the CLI. Check details on logfile %s\n", logger.LogFile)
		log.Println(err)
	}
	log.Printf("debug: cli information dump: %v\n", i)

	switch decision {
	case selfupdate.MustUpdate:
		fmt.Printf("Warning: your current version (%s) is not supported anymore (minimal: %s, latest: %s). You need to update it.\n",
			version.Version, i.MinimalRequiredVersion.String(), i.LatestVersion.String())

		confirmAndUpdate(decision, i, gh)

	case selfupdate.CanUpdate:
		fmt.Printf("Warning: there's a newer version (%s), but this version (%s) is still usable. You can update it by running %s self-update.\n",
			i.LatestVersion.String(), version.Version, os.Args[0])

		// UX decision: do not ask the user for update if it is not required. Just warns.
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
func confirmAndUpdate(d selfupdate.UpdateDecision, i selfupdate.CliInfo, gh *github.Client) {
	if !flagYes && !askIfUpdate() {
		if d == selfupdate.MustUpdate {
			fmt.Println("Cannot continue without updating. Exiting.")
			os.Exit(int(d))
		} else {
			// Update is optional, let the program continue
			return
		}
	}

	fmt.Println("Downloading and applying latest release ...")
	err := selfupdate.DownloadAndApply(i, gh)
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
