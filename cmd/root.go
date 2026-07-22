package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/madmaxieee/loglit/internal/config"
	"github.com/madmaxieee/loglit/internal/proto"
	"github.com/madmaxieee/loglit/internal/reader"
	"github.com/madmaxieee/loglit/internal/renderer"
	"github.com/madmaxieee/loglit/internal/theme"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var flags struct {
	InputFile  string
	OutputFile string
	AppendMode bool
	Profile    string
	Color      string
}

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

	PreRunE: func(_ *cobra.Command, _ []string) error {
		if flags.Color != "auto" && flags.Color != "always" && flags.Color != "never" {
			return fmt.Errorf("invalid color value %q: must be one of auto, always, or never", flags.Color)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) (runErr error) {
		var profileFile *os.File
		if flags.Profile != "" {
			var err error
			profileFile, err = os.Create(flags.Profile)
			if err != nil {
				return fmt.Errorf("create profile: %w", err)
			}
			err = pprof.StartCPUProfile(profileFile)
			if err != nil {
				return errors.Join(fmt.Errorf("start CPU profile: %w", err), profileFile.Close())
			}
			defer func() {
				pprof.StopCPUProfile()
				if err := profileFile.Close(); err != nil {
					runErr = errors.Join(runErr, fmt.Errorf("close profile: %w", err))
				}
			}()
			defer fmt.Fprintln(cmd.OutOrStdout(), "CPU profiling data written to", flags.Profile)
		}

		cfg := config.GetDefaultConfig()
		th := theme.GetDefaultTheme()

		for _, arg := range args {
			if arg == "" {
				continue
			}
			pattern, err := regexp.Compile(arg)
			if err != nil {
				return fmt.Errorf("invalid regex pattern '%s': %v", arg, err)
			}
			cfg.UserSyntax = append(cfg.UserSyntax, proto.Syntax{
				Group:   "UserPattern",
				Pattern: proto.Pattern{Regexp: pattern},
			})
		}

		renderer, err := renderer.New(cfg, th)
		if err != nil {
			return fmt.Errorf("create renderer: %w", err)
		}

		var inputReader io.Reader
		if flags.InputFile == "" {
			inputReader = cmd.InOrStdin()
		} else {
			file, err := openInputFile(flags.InputFile)
			if err != nil {
				return fmt.Errorf("open input file: %w", err)
			}
			defer func() {
				if err := file.Close(); err != nil {
					runErr = errors.Join(runErr, fmt.Errorf("close input file: %w", err))
				}
			}()
			inputReader = file
		}
		bufferedInput := bufio.NewReader(inputReader)

		// Open the two output channels
		stderr := cmd.ErrOrStderr()
		isStderrTerminal := isTerminal(stderr)
		var coloredOutput *bufio.Writer
		colorsEnabled := flags.Color == "always" || flags.Color == "auto" && isStderrTerminal
		if colorsEnabled {
			coloredOutput = bufio.NewWriter(stderr)
		}

		stdout := cmd.OutOrStdout()
		rawOutput, rawOutputCloser, err := rawOutputWriter(flags.OutputFile, stdout, isTerminal(stdout))
		if err != nil {
			return fmt.Errorf("open output: %w", err)
		}
		defer func() {
			if err := rawOutputCloser.Close(); err != nil {
				runErr = errors.Join(runErr, fmt.Errorf("close output: %w", err))
			}
		}()

		// Set up line buffer reader. The command goroutine owns the line buffer
		// and output writers; all events are handled in the loop below.
		chunkCh := reader.ReadChunks(bufferedInput)
		lb := reader.NewLineBuffer(renderer)

		// End gracefully on signal
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(signalCh)

		// Flush periodically to ensure timely output for real-time streams, only when reading from stdin
		var tickerCh <-chan time.Time
		if flags.InputFile == "" {
			ticker := time.NewTicker(500 * time.Millisecond)
			tickerCh = ticker.C
			defer ticker.Stop()
		}

		// Main log processing loop
	inputLoop:
		for {
			select {
			case chunk, ok := <-chunkCh:
				if !ok {
					break inputLoop
				}
				lb.Append(chunk)
				if err := lb.ProcessCompleteLines(coloredOutput, rawOutput); err != nil {
					return fmt.Errorf("write output: %w", err)
				}
			case <-tickerCh:
				if err := lb.FlushPending(coloredOutput, rawOutput); err != nil {
					return fmt.Errorf("flush output: %w", err)
				}
				if coloredOutput != nil {
					if err := coloredOutput.Flush(); err != nil {
						return fmt.Errorf("flush output: %w", err)
					}
				}
				if err := rawOutput.Flush(); err != nil {
					return fmt.Errorf("flush output: %w", err)
				}
			case <-signalCh:
				// Interrupts retain best-effort behavior: flush what is pending,
				// but do not turn cleanup failures into asynchronous errors.
				_ = lb.FlushPending(coloredOutput, rawOutput)
				if coloredOutput != nil {
					_ = coloredOutput.Flush()
				}
				_ = rawOutput.Flush()
				return nil
			}
		}

		var finalizeErr error
		if err := lb.Finalize(coloredOutput, rawOutput); err != nil {
			finalizeErr = fmt.Errorf("write output: %w", err)
		}
		var coloredErr error
		if coloredOutput != nil {
			coloredErr = coloredOutput.Flush()
		}
		rawErr := rawOutput.Flush()
		return errors.Join(finalizeErr, coloredErr, rawErr)
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
	rootCmd.Flags().StringVar(&flags.Color, "color", "auto", "Color mode: auto, always, or never")
	if err := rootCmd.RegisterFlagCompletionFunc("color", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"auto", "always", "never"}, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}
}
