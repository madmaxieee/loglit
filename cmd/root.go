/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"os"

	"github.com/madmaxieee/loglit/internal/config"
	"github.com/madmaxieee/loglit/internal/renderer"
	"github.com/madmaxieee/loglit/internal/theme"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "loglit",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetDefaultConfig()
		th := theme.GetDefaultTheme()

		renderer, err := renderer.New(cfg, th)
		if err != nil {
			// TODO: handle errors properly
			panic(err)
		}

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			coloredLine, err := renderer.Render(line)
			if err != nil {
				panic(err)
			}
			println(coloredLine)
		}

		// TODO: if stdout is piped, write stdin to stdout
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
