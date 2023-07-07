// Copyright © 2016-2021 Wei Shen <shenwei356@gmail.com>
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
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "csvtk",
	Short: "A cross-platform, efficient and practical CSV/TSV toolkit",
	Long: fmt.Sprintf(`csvtk -- a cross-platform, efficient and practical CSV/TSV toolkit

Version: %s

Author: Wei Shen <shenwei356@gmail.com>

Documents  : http://shenwei356.github.io/csvtk
Source code: https://github.com/shenwei356/csvtk

Attention:

  1. The CSV parser requires all the lines have same number of fields/columns.
     Even lines with spaces will cause error. 
     Use '-I/--ignore-illegal-row' to skip these lines if neccessary.
     You can also use 'csvtk fix' to fix files with different numbers of columns in rows.
  2. By default, csvtk thinks your files have header row, if not, switch flag "-H" on.
  3. Column names better be unique.
  4. By default, lines starting with "#" will be ignored, if the header row
     starts with "#", please assign flag "-C" another rare symbol, e.g. '$'.
  5. By default, csvtk handles CSV files, use flag "-t" for tab-delimited files.
  6. If double quotes exist in fields, use flag "-l".
  7. Do not mix use field (column) numbers and names.

Environment variables for frequently used global flags:

  - "CSVTK_T" for flag "-t/--tabs"
  - "CSVTK_H" for flag "-H/--no-header-row"

You can also create a soft link named "tsvtk" for "csvtk", 
which sets "-t/--tabs" by default.

`, VERSION),
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().IntP("chunk-size", "c", 50, `chunk size of CSV reader`)
	RootCmd.PersistentFlags().IntP("num-cpus", "j", runtime.NumCPU(), `number of CPUs to use (default value depends on your computer)`)

	RootCmd.PersistentFlags().StringP("delimiter", "d", ",", `delimiting character of the input CSV file`)
	RootCmd.PersistentFlags().StringP("out-delimiter", "D", ",", `delimiting character of the output CSV file, e.g., -D $'\t' for tab`)
	// RootCmd.PersistentFlags().StringP("quote-char", "q", `"`, `character used to quote strings in the input CSV file`)
	RootCmd.PersistentFlags().StringP("comment-char", "C", `#`, "lines starting with commment-character will be ignored. "+
		`if your header row starts with '#', please assign "-C" another rare symbol, e.g. '$'`)
	RootCmd.PersistentFlags().BoolP("lazy-quotes", "l", false, `if given, a quote may appear in an unquoted field and a non-doubled quote may appear in a quoted field`)

	RootCmd.PersistentFlags().BoolP("tabs", "t", false, `specifies that the input CSV file is delimited with tabs. Overrides "-d"`)
	RootCmd.PersistentFlags().BoolP("out-tabs", "T", false, `specifies that the output is delimited with tabs. Overrides "-D"`)
	RootCmd.PersistentFlags().BoolP("no-header-row", "H", false, `specifies that the input CSV file does not have header row`)
	RootCmd.PersistentFlags().StringP("out-file", "o", "-", `out file ("-" for stdout, suffix .gz for gzipped out)`)

	RootCmd.PersistentFlags().BoolP("ignore-empty-row", "E", false, `ignore empty rows`)
	RootCmd.PersistentFlags().BoolP("ignore-illegal-row", "I", false, `ignore illegal rows. You can also use 'csvtk fix' to fix files with different numbers of columns in rows`)
	RootCmd.PersistentFlags().StringP("infile-list", "", "", "file of input files list (one file per line), if given, they are appended to files from cli arguments")

	RootCmd.CompletionOptions.DisableDefaultCmd = true
	RootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	RootCmd.SetUsageTemplate(usageTemplate(""))
}

func usageTemplate(s string) string {
	return fmt.Sprintf(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}} %s{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsagesWrapped 110 | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsagesWrapped 110 | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`, s)
}
