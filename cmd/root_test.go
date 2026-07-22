package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func resetCLIState(t *testing.T) {
	t.Helper()
	oldFlags := flags
	oldPatterns := patternsFromArgs
	oldIsTerminal := isTerminal
	t.Cleanup(func() {
		flags = oldFlags
		patternsFromArgs = oldPatterns
		isTerminal = oldIsTerminal
		rootCmd.Flags().Set("input", oldFlags.InputFile)
		rootCmd.Flags().Set("output", oldFlags.OutputFile)
		rootCmd.Flags().Set("append", boolString(oldFlags.AppendMode))
		rootCmd.Flags().Set("profile", oldFlags.Profile)
	})
	flags = struct {
		InputFile  string
		OutputFile string
		AppendMode bool
		Profile    string
	}{}
	patternsFromArgs = nil
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func TestCLIInputFileReading(t *testing.T) {
	resetCLIState(t)
	path := filepath.Join(t.TempDir(), "input.log")
	if err := os.WriteFile(path, []byte("one\ntwo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	file, err := openInputFile(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	got, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "one\ntwo\n" {
		t.Fatalf("input = %q", got)
	}
}

func TestCLIRawStdoutRouting(t *testing.T) {
	resetCLIState(t)
	var stdout bytes.Buffer
	isTerminal = func(io.Writer) bool { return false }
	writer, closer, err := rawOutputWriter("", &stdout, isTerminal(&stdout))
	if err != nil {
		t.Fatal(err)
	}
	defer closer.Close()
	_, _ = writer.WriteString("raw output")
	_ = writer.Flush()
	if stdout.String() != "raw output" {
		t.Fatalf("stdout = %q", stdout.String())
	}

	stdout.Reset()
	isTerminal = func(io.Writer) bool { return true }
	writer, closer, err = rawOutputWriter("", &stdout, isTerminal(&stdout))
	if err != nil {
		t.Fatal(err)
	}
	defer closer.Close()
	_, _ = writer.WriteString("hidden")
	_ = writer.Flush()
	if stdout.Len() != 0 {
		t.Fatalf("terminal stdout = %q, want empty", stdout.String())
	}
}

func TestCLIOutputOverwriteAndAppend(t *testing.T) {
	resetCLIState(t)
	path := filepath.Join(t.TempDir(), "output.log")
	if err := os.WriteFile(path, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	flags.AppendMode = false
	writer, closer, err := rawOutputWriter(path, io.Discard, false)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = writer.WriteString("new")
	_ = writer.Flush()
	_ = closer.Close()
	got, _ := os.ReadFile(path)
	if string(got) != "new" {
		t.Fatalf("overwrite output = %q", got)
	}

	flags.AppendMode = true
	writer, closer, err = rawOutputWriter(path, io.Discard, false)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = writer.WriteString("+")
	_ = writer.Flush()
	_ = closer.Close()
	got, _ = os.ReadFile(path)
	if string(got) != "new+" {
		t.Fatalf("append output = %q", got)
	}
}

func TestCLIInvalidRegex(t *testing.T) {
	resetCLIState(t)
	err := rootCmd.Args(rootCmd, []string{"["})
	if err == nil {
		t.Fatal("invalid regex unexpectedly accepted")
	}
	if patternsFromArgs != nil {
		t.Fatalf("invalid regex changed patterns: %v", patternsFromArgs)
	}
}

func TestCLIFileOpenErrors(t *testing.T) {
	resetCLIState(t)
	if _, err := openInputFile(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing input unexpectedly opened")
	}
	if _, err := openOutputFile(filepath.Join(t.TempDir(), "missing", "output"), false); err == nil {
		t.Fatal("invalid output path unexpectedly opened")
	}
}
