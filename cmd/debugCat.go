/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"compress/zlib"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// debugCatCmd represents the debugCat command
var debugCatCmd = &cobra.Command{
	Use:   "debug-cat",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		file, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer file.Close()

		zr, err := zlib.NewReader(file)
		if err != nil {
			return err
		}
		defer zr.Close()

		io.Copy(os.Stdout, zr)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(debugCatCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// debugCatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// debugCatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
