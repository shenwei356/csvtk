// Copyright © 2016-2023 Wei Shen <shenwei356@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
    "github.com/blang/semver"
    "github.com/rhysd/go-github-selfupdate/selfupdate"
)

// VERSION of csvtk
const VERSION = "0.30.0"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version information and check for update",
	Long: `print version information and check for update

`,
	Run: func(cmd *cobra.Command, args []string) {
		app := "csvtk"
		fmt.Printf("%s v%s\n", app, VERSION)

		if !getFlagBool(cmd, "check-update") {
			return
		}

		fmt.Println("\nChecking new version...")
        v := semver.MustParse(VERSION)
        latest, err := selfupdate.UpdateSelf(v, "shenwei356/csvtk")
        if err != nil {
			fmt.Println("Binary update failed:", err)
			return
		}
		if latest.Version.Equals(v) {
			// latest version is the same as current version. It means current binary is up to date.
			fmt.Println("Current binary is the latest version", VERSION)
		} else {
			fmt.Println("Successfully updated to version", latest.Version)
		}
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolP("check-update", "u", false, `check update`)
}
