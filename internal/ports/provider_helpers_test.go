package ports

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// Run the test binary as a subprocess so command behavior is independent of
// installed system tools, shell syntax, and the host platform.
func TestCommandProcess(t *testing.T) {
	if os.Getenv("WHICHPORT_TEST_COMMAND") != "1" {
		return
	}
	args := os.Args[len(os.Args)-3:]
	fmt.Fprint(os.Stdout, args[0])
	fmt.Fprint(os.Stderr, args[1])
	code, _ := strconv.Atoi(args[2])
	os.Exit(code)
}

func TestRunCommandExitHandling(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("WHICHPORT_TEST_COMMAND", "1")
	for _, tt := range []struct {
		name           string
		allowNoMatches bool
		code           int
		stdout         string
		stderr         string
		wantError      bool
	}{
		{name: "empty success"},
		{name: "output success", stdout: "listener"},
		{name: "strict status one", code: 1, wantError: true},
		{name: "strict diagnostic", code: 1, stderr: "netlink failed", wantError: true},
		{name: "no matches", allowNoMatches: true, code: 1},
		{name: "partial matches", allowNoMatches: true, code: 1, stdout: "listener"},
		{name: "diagnostic is not no matches", allowNoMatches: true, code: 1, stderr: "invalid option", wantError: true},
		{name: "other failure", allowNoMatches: true, code: 2, wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			run := runCommand
			if tt.allowNoMatches {
				run = runCommandAllowNoMatches
			}
			got, err := run(context.Background(), executable, "-test.run=^TestCommandProcess$", "--", tt.stdout, tt.stderr, strconv.Itoa(tt.code))
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, wantError = %v", err, tt.wantError)
			}
			if !tt.wantError && got != tt.stdout {
				t.Fatalf("output = %q, want %q", got, tt.stdout)
			}
			if tt.wantError {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) || exitErr.ExitCode() != tt.code {
					t.Fatalf("exit error was not preserved: %v", err)
				}
				if !strings.Contains(err.Error(), tt.stderr) {
					t.Fatalf("diagnostic missing: %v", err)
				}
			}
		})
	}
}

func TestUniquePIDsSkipsUnknownOwners(t *testing.T) {
	got := uniquePIDs([]partialListener{{PID: 0}, {PID: 42}, {PID: 42}, {PID: 7}})
	if want := []int{7, 42}; !reflect.DeepEqual(got, want) {
		t.Fatalf("uniquePIDs() = %v, want %v", got, want)
	}
}
