/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/madmaxieee/loglit/internal/config"
	"github.com/madmaxieee/loglit/internal/proto"
	"github.com/madmaxieee/loglit/internal/renderer"
	"github.com/madmaxieee/loglit/internal/theme"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var flags struct {
	// Define flags here
	InputFile string
}

var patternsFromArgs []regexp.Regexp

var rootCmd = &cobra.Command{
	Use:   "loglit",
	Short: "Loglit is a CLI tool for syntax highlighting and filtering logs",
	Long: `Loglit reads logs from stdin or a file and applies syntax highlighting
based on built-in patterns and user-provided regex patterns. It is designed
to make log analysis easier in the terminal.`,

	Args: func(cmd *cobra.Command, args []string) error {
		for _, arg := range args {
			pattern, err := regexp.Compile(arg)
			if err != nil {
				return fmt.Errorf("invalid regex pattern '%s': %v", arg, err)
			}
			patternsFromArgs = append(patternsFromArgs, *pattern)
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetDefaultConfig()
		th := theme.GetDefaultTheme()

		for _, pattern := range patternsFromArgs {
			cfg.UserSyntax = append(cfg.UserSyntax, proto.Syntax{
				Group:   "UserPattern",
				Pattern: proto.Pattern{Regexp: &pattern},
			})
		}

		renderer, err := renderer.New(cfg, th)
		if err != nil {
			// TODO: handle errors properly
			panic(err)
		}

		shouldWriteStdout := !term.IsTerminal(int(os.Stdout.Fd()))

		var scanner *bufio.Scanner
		if flags.InputFile == "" {
			scanner = bufio.NewScanner(os.Stdin)
		} else {
			file, err := os.Open(flags.InputFile)
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
	rootCmd.Flags().StringVarP(&flags.InputFile, "input", "i", "", "Input file to read logs from, if not provided, reads from stdin")
}
