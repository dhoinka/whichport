package cli

import (
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
