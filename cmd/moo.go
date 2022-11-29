/*
Copyright © 2022 Diego Morales <dgmorales@gmail.com>

*/
package cmd

import (
	"fmt"
	"strings"

	cowsay "github.com/Code-Hex/Neo-cowsay/v2"
	"github.com/dgmorales/go-cli-selfupdate/version"
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
		version.Check()

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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mooCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mooCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
