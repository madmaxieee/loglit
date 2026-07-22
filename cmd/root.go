package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"

	"github.com/madmaxieee/loglit/internal/config"
	"github.com/madmaxieee/loglit/internal/proto"
	"github.com/madmaxieee/loglit/internal/reader"
	"github.com/madmaxieee/loglit/internal/renderer"
	"github.com/madmaxieee/loglit/internal/theme"
	"github.com/madmaxieee/loglit/internal/utils"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var flags struct {
	InputFile  string
	OutputFile string
	AppendMode bool
	Profile    string
}

var patternsFromArgs []regexp.Regexp

var isTerminal = func(w io.Writer) bool {
	file, ok := w.(*os.File)
	return ok && term.IsTerminal(int(file.Fd()))
}

type closeFunc func() error

func (f closeFunc) Close() error { return f() }

func openInputFile(path string) (*os.File, error) {
	return os.Open(path)
}

func openOutputFile(path string, appendMode bool) (*os.File, error) {
	openFlag := os.O_CREATE | os.O_WRONLY
	if appendMode {
		openFlag |= os.O_APPEND
	} else {
		openFlag |= os.O_TRUNC
	}
	return os.OpenFile(path, openFlag, 0644)
}

func rawOutputWriter(path string, stdout io.Writer, stdoutTerminal bool) (*bufio.Writer, io.Closer, error) {
	if path == "" {
		if stdoutTerminal {
			return bufio.NewWriter(io.Discard), closeFunc(func() error { return nil }), nil
		}
		return bufio.NewWriter(stdout), closeFunc(func() error { return nil }), nil
	}

	file, err := openOutputFile(path, flags.AppendMode)
	if err != nil {
		return nil, nil, err
	}
	return bufio.NewWriter(file), file, nil
}

var rootCmd = &cobra.Command{
	Use:   "loglit",
	Short: "Loglit is a CLI tool for syntax highlighting and filtering logs",
	Long: `Loglit reads logs from stdin or a file and applies syntax highlighting
based on built-in patterns and user-provided regex patterns. It is designed
to make log analysis easier in the terminal.`,

	Args: func(cmd *cobra.Command, args []string) error {
		for _, arg := range args {
			if arg == "" {
				continue
			}
			pattern, err := regexp.Compile(arg)
			if err != nil {
				return fmt.Errorf("invalid regex pattern '%s': %v", arg, err)
			}
			patternsFromArgs = append(patternsFromArgs, *pattern)
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		if flags.Profile != "" {
			f, err := os.Create(flags.Profile)
			if err != nil {
				utils.HandleError(err)
			}
			defer f.Close()
			err = pprof.StartCPUProfile(f)
			if err != nil {
				utils.HandleError(err)
			}
			defer pprof.StopCPUProfile()
			defer println("CPU profiling data written to", flags.Profile)
		}

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
			utils.HandleError(err)
		}

		var inputReader io.Reader
		if flags.InputFile == "" {
			inputReader = cmd.InOrStdin()
		} else {
			file, err := openInputFile(flags.InputFile)
			if err != nil {
				utils.HandleError(err)
			}
			defer file.Close()
			inputReader = file
		}
		bufferedInput := bufio.NewReader(inputReader)

		stderr := cmd.ErrOrStderr()
		coloredOutput := bufio.NewWriter(stderr)
		isStderrTerminal := isTerminal(stderr)

		stdout := cmd.OutOrStdout()
		rawOutput, rawOutputCloser, err := rawOutputWriter(flags.OutputFile, stdout, isTerminal(stdout))
		if err != nil {
			utils.HandleError(err)
		}
		defer rawOutputCloser.Close()

		var outputMu sync.Mutex
		defer func() {
			outputMu.Lock()
			coloredOutput.Flush()
			rawOutput.Flush()
			outputMu.Unlock()
		}()

		chunkCh := reader.ReadChunks(bufferedInput)
		lb := reader.NewLineBuffer(renderer)

		// Flush periodically to ensure timely output for real-time streams, only when reading from stdin
		if flags.InputFile == "" {
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()
			go func() {
				for range ticker.C {
					outputMu.Lock()
					if isStderrTerminal {
						lb.FlushPending(coloredOutput, rawOutput)
					} else {
						lb.FlushPending(nil, rawOutput)
					}
					coloredOutput.Flush()
					rawOutput.Flush()
					outputMu.Unlock()
				}
			}()
		}

		// Handle interrupt signal to flush output before exiting
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			outputMu.Lock()
			if isStderrTerminal {
				lb.FlushPending(coloredOutput, rawOutput)
			} else {
				lb.FlushPending(nil, rawOutput)
			}
			coloredOutput.Flush()
			rawOutput.Flush()
			outputMu.Unlock()
			os.Exit(0)
		}()

		for chunk := range chunkCh {
			outputMu.Lock()
			lb.Append(chunk)
			lb.ProcessCompleteLines(coloredOutput, rawOutput)
			outputMu.Unlock()
		}

		outputMu.Lock()
		lb.Finalize(coloredOutput, rawOutput)
		outputMu.Unlock()
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
	rootCmd.Flags().StringVarP(&flags.OutputFile, "output", "o", "", "Output file to write processed logs to")
	rootCmd.Flags().BoolVarP(&flags.AppendMode, "append", "a", false, "Append to the output file instead of overwriting")
	rootCmd.Flags().StringVar(&flags.Profile, "profile", "", "Enable profiling, write CPU profile data to the specified file")
}
