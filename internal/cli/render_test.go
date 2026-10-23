package cli

import (
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/dhoinka/whichport/internal/ports"
)

func TestTruncateMiddle(t *testing.T) {
	t.Parallel()

	if got := truncateMiddle("/usr/local/bin/whichport", 16); got != "/usr/l...ichport" {
		t.Fatalf("truncateMiddle() = %q", got)
	}

	if got := truncateMiddle("short", 16); got != "short" {
		t.Fatalf("truncateMiddle() = %q", got)
	}
}

func TestTruncateSuffix(t *testing.T) {
	t.Parallel()

	if got := truncateSuffix("/System/Library/CoreServices/App.app/Contents/MacOS/ControlCenter", 32); got != ".../Contents/MacOS/ControlCenter" {
		t.Fatalf("truncateSuffix() = %q", got)
	}

	if got := truncateSuffix("short", 16); got != "short" {
		t.Fatalf("truncateSuffix() = %q", got)
	}
}

func TestNormalizeInvocation(t *testing.T) {
	t.Parallel()

	t.Run("drops duplicated executable path", func(t *testing.T) {
		t.Parallel()

		got := normalizeInvocation(listenerFixture(
			"jetbrains-toolbox",
			"/Applications/JetBrains Toolbox.app/Contents/MacOS/jetbrains-toolbox",
			"/Applications/JetBrains Toolbox.app/Contents/MacOS/jetbrains-toolbox --wait-for-pid 40909 --update-successful --minimize",
		))
		want := "jetbrains-toolbox --wait-for-pid 40909 --update-successful --minimize"
		if got != want {
			t.Fatalf("normalizeInvocation() = %q, want %q", got, want)
		}
	})

	t.Run("keeps wrapper command lines", func(t *testing.T) {
		t.Parallel()

		got := normalizeInvocation(listenerFixture(
			"node",
			"/Users/daniel/.nvm/versions/node/v24.14.0/bin/node",
			"ng serve --proxy-config proxy.conf.json",
		))
		if got != "ng serve --proxy-config proxy.conf.json" {
			t.Fatalf("normalizeInvocation() = %q", got)
		}
	})
}

func listenerFixture(command, path, commandLine string) ports.Listener {
	return ports.Listener{
		Command:     command,
		Path:        path,
		CommandLine: commandLine,
	}
}

func TestRenderFitsTerminalWidth(t *testing.T) {
	t.Parallel()
	listeners := []ports.Listener{
		{Port: 65535, Protocol: ports.ProtocolTCP, PID: 1234567, Command: "com.docker.backend", CommandLine: strings.Repeat("long invocation ", 20)},
		{Port: 22, Protocol: ports.ProtocolTCP, Command: "<unavailable>", CommandLine: "<unavailable>"},
	}
	renderer := NewRenderer(io.Discard, false)
	for _, width := range []int{1, 3, 12, 30, 40, 41, 50, 60, 72, 80, 120} {
		t.Run(strconv.Itoa(width), func(t *testing.T) {
			output := renderer.renderWidth(listeners, width)
			for _, line := range strings.Split(output, "\n") {
				if got := visibleWidth(line); got > width {
					t.Errorf("line width = %d exceeds terminal width %d: %q", got, width, line)
				}
			}
			if width >= 50 && (!strings.Contains(output, "65535") || !strings.Contains(output, "1234567")) {
				t.Errorf("table lost port or PID: %s", output)
			}
		})
	}
}

func TestRenderUnknownOwner(t *testing.T) {
	t.Parallel()
	output := NewRenderer(io.Discard, false).renderWidth([]ports.Listener{
		{Port: 22, Protocol: ports.ProtocolTCP, Command: "<unavailable>", CommandLine: "<unavailable>"},
	}, 80)
	if !strings.Contains(output, "22    TCP    -") || !strings.Contains(output, "<unavailable>") {
		t.Fatalf("unknown owner was not rendered explicitly: %s", output)
	}
}
