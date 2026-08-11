package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"
)

var errCommandWriter = errors.New("command writer failed")

type commandFailingWriter struct{}

func (commandFailingWriter) Write([]byte) (int, error) { return 0, errCommandWriter }

func resetCLIState(t *testing.T) {
	t.Helper()
	oldFlags := flags
	oldIsTerminal := isTerminal
	oldIn, oldOut, oldErr := rootCmd.InOrStdin(), rootCmd.OutOrStdout(), rootCmd.ErrOrStderr()
	oldVersionFlag := rootCmd.Flags().Lookup("version")
	var oldVersionFlagValue string
	var oldVersionFlagChanged bool
	if oldVersionFlag != nil {
		oldVersionFlagValue = oldVersionFlag.Value.String()
		oldVersionFlagChanged = oldVersionFlag.Changed
	}
	t.Cleanup(func() {
		flags = oldFlags
		isTerminal = oldIsTerminal
		rootCmd.SetIn(oldIn)
		rootCmd.SetOut(oldOut)
		rootCmd.SetErr(oldErr)
		rootCmd.SetArgs(nil)
		if versionFlag := rootCmd.Flags().Lookup("version"); versionFlag != nil {
			if oldVersionFlag == nil {
				_ = versionFlag.Value.Set("false")
				versionFlag.Changed = false
			} else {
				_ = versionFlag.Value.Set(oldVersionFlagValue)
				versionFlag.Changed = oldVersionFlagChanged
			}
		}
		_ = rootCmd.Flags().Set("input", oldFlags.InputFile)
		_ = rootCmd.Flags().Set("output", oldFlags.OutputFile)
		_ = rootCmd.Flags().Set("append", boolString(oldFlags.AppendMode))
		_ = rootCmd.Flags().Set("profile", oldFlags.Profile)
		_ = rootCmd.Flags().Set("color", oldFlags.Color)
		_ = rootCmd.Flags().Set("no-peek", boolString(oldFlags.NoPeek))
	})
	flags = struct {
		InputFile  string
		OutputFile string
		AppendMode bool
		Profile    string
		Color      string
		NoPeek     bool
	}{Color: "auto"}
	isTerminal = func(_ io.Writer) bool { return false }
	if versionFlag := rootCmd.Flags().Lookup("version"); versionFlag != nil {
		_ = versionFlag.Value.Set("false")
		versionFlag.Changed = false
	}
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func executeCLI(t *testing.T, input string, args ...string) (string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	rootCmd.SetIn(bytes.NewBufferString(input))
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	if err != nil && stderr.Len() > 0 {
		t.Logf("stderr: %s", stderr.String())
	}
	return stdout.String(), err
}

func TestRootExecutesWithInjectedNonTTYIO(t *testing.T) {
	resetCLIState(t)
	output, err := executeCLI(t, "one\ntwo\n")
	if err != nil {
		t.Fatal(err)
	}
	if output != "one\ntwo\n" {
		t.Fatalf("output = %q, want input unchanged", output)
	}
}

func TestRootPrintsVersion(t *testing.T) {
	resetCLIState(t)
	oldVersion := version
	oldCommandVersion := rootCmd.Version
	t.Cleanup(func() {
		version = oldVersion
		rootCmd.Version = oldCommandVersion
	})
	version = "test-version"
	rootCmd.Version = version

	output, err := executeCLI(t, "", "--version")
	if err != nil {
		t.Fatal(err)
	}
	if output != "loglit version test-version\n" {
		t.Fatalf("output = %q, want %q", output, "loglit version test-version\n")
	}
}

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		name     string
		settings []debug.BuildSetting
		want     string
	}{
		{
			name: "dirty metadata",
			settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abcdef1234567890"},
				{Key: "vcs.time", Value: "2026-08-11T01:02:03-07:00"},
				{Key: "vcs.modified", Value: "true"},
			},
			want: "dev (abcdef, 2026-08-11, dirty)",
		},
		{
			name: "clean metadata",
			settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abcdef1234567890"},
				{Key: "vcs.time", Value: "2026-08-11T01:02:03Z"},
				{Key: "vcs.modified", Value: "false"},
			},
			want: "dev (abcdef, 2026-08-11)",
		},
		{
			name:     "missing metadata",
			settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "abcdef1234567890"}},
			want:     "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatVersion("dev", &debug.BuildInfo{Settings: tt.settings}); got != tt.want {
				t.Fatalf("formatVersion() = %q, want %q", got, tt.want)
			}
		})
	}

	if got := formatVersion("1.2.3", &debug.BuildInfo{Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "ignored"}}}); got != "1.2.3" {
		t.Fatalf("release version = %q, want linker-injected version unchanged", got)
	}
}

func TestRootSkipsColoredOutputForNonTTYStderr(t *testing.T) {
	resetCLIState(t)
	var stdout, stderr bytes.Buffer
	rootCmd.SetIn(bytes.NewBufferString("INFO\n"))
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs(nil)
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr output = %q, want empty", stderr.String())
	}
}

func TestRootKeepsColoredOutputForTTYStderr(t *testing.T) {
	resetCLIState(t)
	var stdout, stderr bytes.Buffer
	rootCmd.SetIn(bytes.NewBufferString("INFO\n"))
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs(nil)
	isTerminal = func(w io.Writer) bool { return w == &stderr }
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if stdout.String() != "INFO\n" {
		t.Fatalf("stdout output = %q, want raw output", stdout.String())
	}
	if !bytes.Contains(stderr.Bytes(), []byte("\x1b[")) {
		t.Fatalf("stderr output = %q, want ANSI-colored output", stderr.String())
	}
}

func TestRootColorModes(t *testing.T) {
	tests := []struct {
		name       string
		mode       string
		stdoutTTY  bool
		stderrTTY  bool
		stdoutAnsi bool
		stderrAnsi bool
	}{
		{name: "auto non-terminal", mode: "auto"},
		{name: "auto terminal", mode: "auto", stdoutTTY: true, stdoutAnsi: true},
		{name: "always non-terminal", mode: "always", stdoutAnsi: true},
		{name: "never terminal", mode: "never", stdoutTTY: true},
		{name: "peek is independent of color", mode: "never", stderrTTY: true, stderrAnsi: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetCLIState(t)
			var stdout, stderr bytes.Buffer
			rootCmd.SetIn(bytes.NewBufferString("INFO\n"))
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs([]string{"--color", tt.mode})
			isTerminal = func(w io.Writer) bool {
				return (tt.stdoutTTY && w == &stdout) || (tt.stderrTTY && w == &stderr)
			}
			if err := rootCmd.Execute(); err != nil {
				t.Fatal(err)
			}
			if got := bytes.Contains(stdout.Bytes(), []byte("\x1b[")); got != tt.stdoutAnsi {
				t.Fatalf("stdout colored output = %v, stdout = %q; want %v", got, stdout.String(), tt.stdoutAnsi)
			}
			if got := bytes.Contains(stderr.Bytes(), []byte("\x1b[")); got != tt.stderrAnsi {
				t.Fatalf("stderr colored output = %v, stderr = %q; want %v", got, stderr.String(), tt.stderrAnsi)
			}
		})
	}
}

func TestRootNoPeekAndNoAvailableTTY(t *testing.T) {
	for _, noPeek := range []bool{false, true} {
		t.Run(fmt.Sprintf("no-peek=%v", noPeek), func(t *testing.T) {
			resetCLIState(t)
			var stdout, stderr bytes.Buffer
			rootCmd.SetIn(bytes.NewBufferString("INFO\n"))
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			args := []string{}
			if noPeek {
				args = append(args, "--no-peek")
			}
			rootCmd.SetArgs(args)
			isTerminal = func(_ io.Writer) bool { return false }
			if err := rootCmd.Execute(); err != nil {
				t.Fatal(err)
			}
			if stdout.String() != "INFO\n" || stderr.Len() != 0 {
				t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
		})
	}
}

func TestRootNoPeekDisablesAvailableTTY(t *testing.T) {
	resetCLIState(t)
	var stdout, stderr bytes.Buffer
	rootCmd.SetIn(bytes.NewBufferString("INFO\n"))
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"--no-peek"})
	isTerminal = func(w io.Writer) bool { return w == &stderr }
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if stdout.String() != "INFO\n" || stderr.Len() != 0 {
		t.Fatalf("stdout=%q stderr=%q, want raw stdout and no peek", stdout.String(), stderr.String())
	}
}

func TestRootOutputFileIsRawAndPeekIsColored(t *testing.T) {
	resetCLIState(t)
	outputPath := filepath.Join(t.TempDir(), "output.log")
	var stdout, stderr bytes.Buffer
	rootCmd.SetIn(bytes.NewBufferString("INFO\n"))
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"--output", outputPath, "--color", "always"})
	isTerminal = func(w io.Writer) bool { return w == &stdout }
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "INFO\n" {
		t.Fatalf("file output = %q, want uncolored primary output", contents)
	}
	if stdout.String() == "" || !bytes.Contains(stdout.Bytes(), []byte("\x1b[")) || stderr.Len() != 0 {
		t.Fatalf("stdout=%q stderr=%q, want colored stdout peek only", stdout.String(), stderr.String())
	}
}

func TestRootRejectsInvalidColorMode(t *testing.T) {
	resetCLIState(t)
	_, err := executeCLI(t, "INFO\n", "--color", "sometimes")
	if err == nil || err.Error() != `invalid color value "sometimes": must be one of auto, always, or never` {
		t.Fatalf("error = %v, want clear invalid color error", err)
	}
}

func TestRootReadsFileInput(t *testing.T) {
	resetCLIState(t)
	inputPath := filepath.Join(t.TempDir(), "input.log")
	if err := os.WriteFile(inputPath, []byte("from file\n"), 0644); err != nil {
		t.Fatal(err)
	}
	output, err := executeCLI(t, "", "--input", inputPath)
	if err != nil {
		t.Fatal(err)
	}
	if output != "from file\n" {
		t.Fatalf("output = %q", output)
	}
}

func TestRootOutputOverwriteAndAppend(t *testing.T) {
	resetCLIState(t)
	outputPath := filepath.Join(t.TempDir(), "output.log")
	if err := os.WriteFile(outputPath, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := executeCLI(t, "new\n", "--output", outputPath); err != nil {
		t.Fatal(err)
	}
	if _, err := executeCLI(t, "+\n", "--output", outputPath, "--append"); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "new\n+\n" {
		t.Fatalf("output file = %q", contents)
	}
}

func TestRootRejectsInvalidRegex(t *testing.T) {
	resetCLIState(t)
	_, err := executeCLI(t, "input\n", "[")
	if err == nil {
		t.Fatal("invalid regex unexpectedly accepted")
	}
}

func TestRootReportsFileOpenFailures(t *testing.T) {
	resetCLIState(t)
	_, err := executeCLI(t, "", "--input", filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("missing input unexpectedly succeeded")
	}

	resetCLIState(t)
	_, err = executeCLI(t, "input\n", "--output", filepath.Join(t.TempDir(), "missing", "output"))
	if err == nil {
		t.Fatal("invalid output path unexpectedly succeeded")
	}
}

func TestRootReportsInjectedWriterFailure(t *testing.T) {
	resetCLIState(t)
	rootCmd.SetIn(bytes.NewBufferString("line\n"))
	rootCmd.SetOut(commandFailingWriter{})
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs(nil)
	if err := rootCmd.Execute(); err == nil || !errors.Is(err, errCommandWriter) {
		t.Fatalf("error = %v, want %v", err, errCommandWriter)
	}
}

func FuzzRootAcceptsArbitraryInput(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("one\ntwo\n"))
	f.Add([]byte{0xff, 0xfe, 0x00, 0x80})
	f.Add([]byte("unterminated line"))

	f.Fuzz(func(t *testing.T, data []byte) {
		resetCLIState(t)
		_, _ = executeCLI(t, string(data))
	})
}
