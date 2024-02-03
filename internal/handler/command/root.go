package command

import (
	"fmt"
	"github.com/fanky5g/ponzu/internal/handler/command/generate"
	"github.com/fanky5g/ponzu/internal/handler/command/serve"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:  "ponzu",
	Long: `Ponzu is an open-source HTTP server framework and CMS`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(RootCmd.ErrOrStderr(), err)
		os.Exit(1)
	}
}

func init() {
	serve.RegisterCommandRecursive(RootCmd)
	generate.RegisterCommandRecursive(RootCmd)
}
