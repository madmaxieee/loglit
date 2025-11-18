/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/madmaxieee/loglit/internal/config"
	"github.com/madmaxieee/loglit/internal/renderer"
	"github.com/madmaxieee/loglit/internal/theme"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var rootCmd = &cobra.Command{
	Use:   "loglit",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// TODO: use args to take user specified highlight patterns, use a flag to decide whether or not to open a file
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetDefaultConfig()
		th := theme.GetDefaultTheme()

		renderer, err := renderer.New(cfg, th)
		if err != nil {
			// TODO: handle errors properly
			panic(err)
		}

		shouldWriteStdout := !term.IsTerminal(int(os.Stdout.Fd()))

		var scanner *bufio.Scanner
		if len(args) == 0 {
			scanner = bufio.NewScanner(os.Stdin)
		} else {
			file, err := os.Open(args[0])
			if err != nil {
				panic(err)
			}
			defer file.Close()
			scanner = bufio.NewScanner(file)
		}

		for scanner.Scan() {
			line := scanner.Text()
			coloredLine, err := renderer.Render(line)
			if err != nil {
				panic(err)
			}
			// writes coloredLine to stderr
			fmt.Fprintln(os.Stderr, coloredLine)
			if shouldWriteStdout {
				// write original line to stdout
				fmt.Fprintln(os.Stdout, line)
			}
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
