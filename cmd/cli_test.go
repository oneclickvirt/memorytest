package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestParseCLIOptions(t *testing.T) {
	opts, err := parseCLI([]string{"--json", "--size", "1048576", "--iterations", "2", "--timeout", "1s"})
	if err != nil {
		t.Fatalf("parseCLI returned error: %v", err)
	}
	if !opts.jsonOutput || opts.sizeBytes != 1048576 || opts.iterations != 2 || opts.timeout != time.Second {
		t.Fatalf("unexpected options: %#v", opts)
	}
}

func TestHelpRetainsLegacyFlags(t *testing.T) {
	var output bytes.Buffer
	newFlagSet(&cliOptions{}, &output).PrintDefaults()
	for _, legacy := range []string{"-h", "-l string", "-m string", "-log", "-v"} {
		if !strings.Contains(output.String(), legacy) {
			t.Fatalf("help is missing legacy flag %q: %s", legacy, output.String())
		}
	}
}

func TestParseCLIRejectsNegativeSize(t *testing.T) {
	if _, err := parseCLI([]string{"--size", "-1"}); err == nil {
		t.Fatal("expected negative size to be rejected")
	}
}

func TestParseCLIRejectsInvalidAndIgnoredOptions(t *testing.T) {
	for _, args := range [][]string{
		{"-l", "fr"},
		{"-m", "unknown"},
		{"--size", "1048576"},
		{"--structured", "-m", "stream"},
		{"--structured", "--size", "0"},
		{"--structured", "--iterations", "21"},
		{"--structured", "--timeout", "0s"},
		{"unexpected"},
	} {
		if _, err := parseCLI(args); err == nil {
			t.Fatalf("expected arguments %v to be rejected", args)
		}
	}
}

func TestParseCLIHelpAndVersionIgnoreOtherInvalidValues(t *testing.T) {
	for _, args := range [][]string{{"-h", "-l", "fr"}, {"-v", "-m", "unknown"}} {
		if _, err := parseCLI(args); err != nil {
			t.Fatalf("help/version arguments %v returned %v", args, err)
		}
	}
}

func TestCLIActionPrioritizesHelpAndVersion(t *testing.T) {
	if got := selectCLIAction(cliOptions{help: true, version: true, jsonOutput: true}); got != "help" {
		t.Fatalf("help action = %q", got)
	}
	if got := selectCLIAction(cliOptions{version: true, jsonOutput: true}); got != "version" {
		t.Fatalf("version action = %q", got)
	}
}
