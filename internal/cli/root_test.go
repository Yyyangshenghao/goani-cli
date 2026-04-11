package cli

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunWithoutArgsPrintsTopLevelUsage(t *testing.T) {
	output := runCLIAndCaptureOutput(t, "goani")

	assertContains(t, output, "goani - 命令行动漫播放器")
	assertContains(t, output, "config")
	assertContains(t, output, "search")
	assertContains(t, output, "source")
	assertContains(t, output, "tui")
	assertContains(t, output, "version")
	assertNotContains(t, output, "proxy-hls")
}

func TestRunHelpConfigPrintsConfigUsage(t *testing.T) {
	output := runCLIAndCaptureOutput(t, "goani", "help", "config")

	assertContains(t, output, "goani config player <name> <path>")
	assertContains(t, output, "goani config player default <name>")
}

func TestRunSourceHelpPrintsSourceUsage(t *testing.T) {
	output := runCLIAndCaptureOutput(t, "goani", "source", "--help")

	assertContains(t, output, "goani source list")
	assertContains(t, output, "goani source refresh")
	assertContains(t, output, "goani source reset")
}

func TestRunVersionAliasPrintsVersionInfo(t *testing.T) {
	output := runCLIAndCaptureOutput(t, "goani", "--version")

	assertContains(t, output, "goani ")
	assertContains(t, output, "Git commit:")
	assertContains(t, output, "Build date:")
}

func runCLIAndCaptureOutput(t *testing.T, args ...string) string {
	t.Helper()

	oldArgs := os.Args
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}

	os.Args = args
	os.Stdout = w
	os.Stderr = w

	Run()

	_ = w.Close()
	os.Args = oldArgs
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	_ = r.Close()

	return string(output)
}

func assertContains(t *testing.T, output, want string) {
	t.Helper()
	if !strings.Contains(output, want) {
		t.Fatalf("expected output to contain %q, got:\n%s", want, output)
	}
}

func assertNotContains(t *testing.T, output, unwanted string) {
	t.Helper()
	if strings.Contains(output, unwanted) {
		t.Fatalf("expected output not to contain %q, got:\n%s", unwanted, output)
	}
}
