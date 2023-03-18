/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/mizuho-u/got/usecase"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "got",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	SilenceUsage: true,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.got.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringP("path", "C", "", "path")
}

const gotdir string = ".git"

func newContext(workspace string, cmd *cobra.Command) (usecase.GotContext, error) {

	if workspace == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		workspace = wd
	}

	if !filepath.IsAbs(workspace) {

		abs, err := filepath.Abs(workspace)
		if err != nil {
			return nil, err
		}

		workspace = abs
	}

	return usecase.NewContext(context.Background(), workspace, gotdir, cmd.OutOrStdout(), cmd.OutOrStderr()), nil
}
