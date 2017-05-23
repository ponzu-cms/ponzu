package main

import (
	"fmt"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/spf13/cobra"
)

var templateFuncs = template.FuncMap{
	"rpad": rpad,
	"trimTrailingWhitespaces": trimRightSpace,
}

var tmpl = `{{with (or .Cmd.Long .Cmd.Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Cmd.Runnable .Cmd.HasSubCommands}}Usage:{{if .Cmd.Runnable}}
  {{.Cmd.UseLine}}{{end}}{{if .Cmd.HasAvailableSubCommands}}
  {{.Cmd.CommandPath}} [command]{{end}}{{if gt (len .Cmd.Aliases) 0}}

Aliases:
  {{.Cmd.NameAndAliases}}{{end}}{{if .Cmd.HasExample}}

Examples:
{{.Cmd.Example}}{{end}}{{if .Cmd.HasAvailableSubCommands}}

Available Commands:{{range .Cmd.Commands}}{{if (or .IsAvailableCommand false)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .Cmd.HasAvailableLocalFlags}}

Flags for all commands:
{{.Cmd.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{range .Subs}}{{if (and .IsAvailableCommand .HasAvailableLocalFlags)}}

Flags for '{{.Name}}' command:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{end}}{{if .Cmd.HasHelpSubCommands}}

Additional help topics:{{range .Cmd.Commands}}{{if .Cmd.IsAdditionalHelpTopicCommand}}
  {{rpad .Cmd.CommandPath .Cmd.CommandPathPadding}} {{.Cmd.Short}}{{end}}{{end}}{{end}}{{if .Cmd.HasAvailableSubCommands}}

Use "{{.Cmd.CommandPath}} [command] --help" for more information about a command.{{end}}
{{end}}`

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "help about any command",
	Run: func(cmd *cobra.Command, args []string) {
		cmd, _, e := rootCmd.Find(args)
		if cmd == nil || e != nil {
			rootCmd.Printf("Unknown help topic %#q\n", args)
			rootCmd.Usage()
			return
		}
		t := template.New("help")
		t.Funcs(templateFuncs)
		template.Must(t.Parse(tmpl))
		if len(args) > 0 {
			rootCmd.HelpFunc()(cmd, args)
			return
		}

		sortByName := func(i, j int) bool { return cmds[i].Name() < cmds[j].Name() }
		sort.Slice(cmds, sortByName)

		err := t.Execute(cmd.OutOrStdout(), struct {
			Cmd  *cobra.Command
			Subs []*cobra.Command
		}{
			Cmd:  rootCmd,
			Subs: cmds})
		if err != nil {
			cmd.Println(err)
		}
	},
}

var cmds []*cobra.Command

// RegisterCmdlineCommand adds a cobra command to the root command and makes it
// known to the main package
func RegisterCmdlineCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
	cmds = append(cmds, cmd)
}

func init() {
	rootCmd.AddCommand(helpCmd)
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}

func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}
