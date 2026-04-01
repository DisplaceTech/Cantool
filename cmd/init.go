package cmd

import (
	"fmt"

	"github.com/displacetech/cantool/internal/output"
	"github.com/displacetech/cantool/internal/scaffold"
	"github.com/spf13/cobra"
)

var templateFlag string

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Create a new Canton project",
	Long:  "Scaffold a new Canton project directory with DAML sources, tests, and configuration files.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInit,
}

func init() {
	initCmd.Flags().StringVarP(&templateFlag, "template", "t", "basic", `Project template: "basic" or "token"`)
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	out := output.New(Format())

	if len(args) == 0 {
		err := output.Errorf("CT1001", "Provide a project name: cantool init <name>", "project name is required")
		out.Error(err)
		return err
	}

	name := args[0]

	if templateFlag != "basic" && templateFlag != "token" {
		err := output.Errorf("CT1002", `Use --template basic or --template token`, "unknown template %q", templateFlag)
		out.Error(err)
		return err
	}

	if err := scaffold.Scaffold(name, templateFlag, "."); err != nil {
		cantoolErr := output.Errorf("CT1003", "Check the project name and target directory", "scaffold failed: %s", err.Error())
		out.Error(cantoolErr)
		return cantoolErr
	}

	out.Success(fmt.Sprintf("Created project %q with %q template", name, templateFlag), nil)
	out.Info("")
	out.Info("Next steps:")
	out.Info(fmt.Sprintf("  cd %s", name))
	out.Info("  cantool build     # compile DAML sources")
	out.Info("  cantool test      # run DAML test scripts")
	out.Info("  cantool sandbox   # start a local sandbox")

	return nil
}
