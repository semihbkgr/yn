package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/semihbkgr/yn/model"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yn",
	Short: "yn yaml navigator",
	Long: `yn
yaml navigator
yn < file.yaml
`,
	SilenceUsage: true,
	RunE:         run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP(model.FlagFile, "f", "", "input file")
}

func run(cmd *cobra.Command, args []string) error {
	opts, err := model.NewOptions(cmd)
	if err != nil {
		return err
	}

	output, err := model.RunProgram(cmd.Context(), opts)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(cmd.OutOrStdout(), output)
	return err
}
