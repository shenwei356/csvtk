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

  1. By default, csvtk assumes input files have header row, if not, switch flag "-H" on.
  2. By default, csvtk handles CSV files, use flag "-t" for tab-delimited files.
  3. Column names should be unique.
  4. By default, lines starting with "#" will be ignored, if the header row
     starts with "#", please assign flag "-C" another rare symbol, e.g. '$'.
  5. Do not mix use field (column) numbers and names to specify columns to operate.
  6. The CSV parser requires all the lines have same numbers of fields/columns.
     Even lines with spaces will cause error.
     Use '-I/--ignore-illegal-row' to skip these lines if neccessary.
     You can also use "csvtk fix" to fix files with different numbers of columns in rows.
  7. If double-quotes exist in fields not enclosed with double-quotes, e.g.,
         x,a "b" c,1
     It would report error:
         bare " in non-quoted-field.
     Please switch on the flag "-l" or use "csvtk fix-quotes" to fix it.
  8. If somes fields have only a double-quote eighter in the beginning or in the end, e.g.,
         x,d "e","a" b c,1
     It would report error:
         extraneous or missing " in quoted-field
     Please use "csvtk fix-quotes" to fix it, and use "csvtk del-quotes" to reset to the
     original format as needed.
  9. csvtk writes gzip files very fast, much faster than the multi-threaded pigz,
     therefore there's no need to pipe the result to gzip/pigz.
     csvtk also supports reading and writing xz (.xz), zstd (.zst) and Bzip2 (.bz2) formats.

Environment variables for frequently used global flags:

  - "CSVTK_T" for flag "-t/--tabs"
  - "CSVTK_H" for flag "-H/--no-header-row"
  - "CSVTK_QUIET" for flag "--quiet"

You can also create a symbolic link named "tsvtk" to csvtk, or simply create a copy
of the executable called "tsvtk". When invoked as "tsvtk", csvtk will automatically
enable the "-t/--tabs" flag.

`, VERSION),

	Run: func(cmd *cobra.Command, args []string) {
		if getFlagBool(cmd, "version") {
			fmt.Printf("csvtk v%s\n", VERSION)
			os.Exit(0)
		} else {
			cmd.Help()
		}
	},
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
	RootCmd.AddGroup(
		&cobra.Group{
			ID:    "info",
			Title: "Commands for Information:",
		},
		&cobra.Group{
			ID:    "format",
			Title: "Format Conversion:",
		},
		&cobra.Group{
			ID:    "set",
			Title: "Commands for Set Operation:",
		},
		&cobra.Group{
			ID:    "edit",
			Title: "Commands for Edit:",
		},
		&cobra.Group{
			ID:    "transform",
			Title: "Commands for Data Transformation:",
		},
		&cobra.Group{
			ID:    "order",
			Title: "Commands for Ordering:",
		},
		&cobra.Group{
			ID:    "plot",
			Title: "Commands for Ploting:",
		},
		&cobra.Group{
			ID:    "misc",
			Title: "Commands for Miscellaneous Functions:",
		},
	)

	defaultThreads := runtime.NumCPU()
	if defaultThreads > 4 {
		defaultThreads = 4
	}
	RootCmd.PersistentFlags().IntP("num-cpus", "j", defaultThreads, `number of CPUs to use`)

	RootCmd.PersistentFlags().BoolP("quiet", "", false, "be quiet and do not show extra information and warnings")

	RootCmd.PersistentFlags().StringP("delimiter", "d", ",", `delimiting character of the input CSV file`)
	RootCmd.PersistentFlags().StringP("out-delimiter", "D", ",", `delimiting character of the output CSV file, e.g., -D $'\t' for tab`)
	// RootCmd.PersistentFlags().StringP("quote-char", "q", `"`, `character used to quote strings in the input CSV file`)
	RootCmd.PersistentFlags().StringP("comment-char", "C", `#`, "lines starting with commment-character will be ignored. "+
		`if your header row starts with '#', please assign "-C" another rare symbol, e.g. '$'`)
	RootCmd.PersistentFlags().BoolP("lazy-quotes", "l", false, `if given, a quote may appear in an unquoted field and a non-doubled quote may appear in a quoted field`)

	RootCmd.PersistentFlags().BoolP("tabs", "t", false, `specifies that the input CSV file is delimited with tabs. Overrides "-d"`)
	RootCmd.PersistentFlags().BoolP("out-tabs", "T", false, `specifies that the output is delimited with tabs. Overrides "-D"`)
	RootCmd.PersistentFlags().BoolP("no-header-row", "H", false, `specifies that the input CSV file does not have header row`)
	RootCmd.PersistentFlags().BoolP("delete-header", "U", false, `do not output header row`)
	RootCmd.PersistentFlags().StringP("out-file", "o", "-", `out file ("-" for stdout, suffix .gz for gzipped out)`)

	RootCmd.PersistentFlags().BoolP("show-row-number", "Z", false, `show row number as the first column, with header row skipped`)

	RootCmd.PersistentFlags().BoolP("ignore-empty-row", "E", false, `ignore empty rows`)
	RootCmd.PersistentFlags().BoolP("ignore-illegal-row", "I", false, `ignore illegal rows. You can also use 'csvtk fix' to fix files with different numbers of columns in rows`)
	RootCmd.PersistentFlags().StringP("infile-list", "X", "", "file of input files list (one file per line), if given, they are appended to files from cli arguments")

	RootCmd.PersistentFlags().BoolP("version", "V", false, "print version information")

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
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsagesWrapped 110 | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsagesWrapped 110 | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`, s)
}
