/*
Copyright © 2022 Diego Morales <dgmorales@gmail.com>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
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
required. Exit code will reflect the version state:

IsLatest   = 0
CanUpdate  = 10
MustUpdate = 20
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

// versionCheck checks if current version can or must be updated, and interacts with the user
// about it
func versionCheck(v version.Checker) version.Assertion {
	ans := v.Check()

	switch ans {
	case version.MustUpdate:
		fmt.Printf("Warning: your current version (%s) is not supported anymore (minimal: %s, latest: %s). You need to update it.\n",
			v.Current(), v.Minimal(), v.Latest())

		if !flagCheck {
			confirmAndUpdate(ans, v)
		}

	case version.CanUpdate:
		fmt.Printf("Warning: there's a newer version (%s), but this version (%s) is still usable. You can update it by running %s self-update.\n",
			v.Latest(), v.Current(), os.Args[0])
		// UX decision: just warns, and do not ask the user for update if it is not required.

	case version.IsBeyond:
		fmt.Printf("Warning: your are using a development version (current %s > latest release %s).\n",
			v.Current(), v.Latest())

	case version.IsUnknown:
		fmt.Println("Warning: couldn't check if you are running the latest (or minimal) version.")
	}

	return ans
}

// confirmAndUpdate will confirms if we can proceed with the self-update,
// and performs the update if confirmed.
//
// It will ask the user, check for --yes flags, etc.
//
// If the update is **not required** and not performed this function returns.
// Otherwise this function ensures the program is terminated.
func confirmAndUpdate(a version.Assertion, v version.Checker) {
	if !flagYes && !askIfUpdate() {
		if a == version.MustUpdate {
			fmt.Println("Cannot continue without updating. Exiting.")
			os.Exit(int(a))
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
