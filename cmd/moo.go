/*
Copyright © 2022 Diego Morales <dgmorales@gmail.com>

*/
package cmd

import (
	"fmt"
	"strings"

	cowsay "github.com/Code-Hex/Neo-cowsay/v2"
	"github.com/dgmorales/go-cli-selfupdate/start"
	"github.com/spf13/cobra"
)

// mooCmd represents the moo command
var mooCmd = &cobra.Command{
	Use:   "moo",
	Short: "Cow-Echo something back to console",
	Long: `I'm a dummy cmd in a test app with other purposes.

Cowsaying is fun. What's more to say?
`,
	Run: func(cmd *cobra.Command, args []string) {
		// we need no state at all on this command
		start.ForLocalUse()

		say := "moo"
		if len(args) > 0 {
			say = strings.Join(args, " ")
		}

		mooSay, err := cowsay.Say(say)

		if err != nil {
			panic(err) // I'm a dummy command, I can allow myself to panic occasionally
		}

		fmt.Println(mooSay)

	},
}

func init() {
	rootCmd.AddCommand(mooCmd)
}
