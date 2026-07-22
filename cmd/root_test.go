package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
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
	t.Cleanup(func() {
		flags = oldFlags
		isTerminal = oldIsTerminal
		rootCmd.SetIn(oldIn)
		rootCmd.SetOut(oldOut)
		rootCmd.SetErr(oldErr)
		rootCmd.SetArgs(nil)
		_ = rootCmd.Flags().Set("input", oldFlags.InputFile)
		_ = rootCmd.Flags().Set("output", oldFlags.OutputFile)
		_ = rootCmd.Flags().Set("append", boolString(oldFlags.AppendMode))
		_ = rootCmd.Flags().Set("profile", oldFlags.Profile)
	})
	flags = struct {
		InputFile  string
		OutputFile string
		AppendMode bool
		Profile    string
	}{}
	isTerminal = func(_ io.Writer) bool { return false }
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
