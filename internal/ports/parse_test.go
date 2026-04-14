package ports

import "testing"

func TestParseSocketOutput(t *testing.T) {
	t.Parallel()

	output := `p1107
cnode
n[::1]:3000
n127.0.0.1:3000
p43267
ccom.docker.backend
n*:5432
`

	got := parseSocketOutput(ProtocolTCP, output)
	if len(got) != 2 {
		t.Fatalf("expected 2 listeners, got %d", len(got))
	}

	if got[0].PID != 1107 || got[0].Port != 3000 || got[0].Command != "node" {
		t.Fatalf("unexpected first listener: %#v", got[0])
	}

	if got[1].PID != 43267 || got[1].Port != 5432 || got[1].Command != "com.docker.backend" {
		t.Fatalf("unexpected second listener: %#v", got[1])
	}
}

func TestParseSocketOutputIgnoresConnectedUDPSockets(t *testing.T) {
	t.Parallel()

	output := `p6059
ccollama
n127.0.0.1:11434
n127.0.0.1:11434->127.0.0.1:55234
`

	got := parseSocketOutput(ProtocolUDP, output)
	if len(got) != 1 {
		t.Fatalf("expected 1 listener, got %d", len(got))
	}
}

func TestParsePathOutput(t *testing.T) {
	t.Parallel()

	output := `p1107
ftxt
n/usr/libexec/rapportd
ftxt
n/System/Library/irrelevant.dylib
p43267
ftxt
n/Applications/Docker.app/Contents/MacOS/com.docker.backend
`

	got := parsePathOutput(output)
	if got[1107] != "/usr/libexec/rapportd" {
		t.Fatalf("unexpected path for 1107: %q", got[1107])
	}

	if got[43267] != "/Applications/Docker.app/Contents/MacOS/com.docker.backend" {
		t.Fatalf("unexpected path for 43267: %q", got[43267])
	}
}

func TestParsePSMetadata(t *testing.T) {
	t.Parallel()

	output := `1107 /usr/libexec/rapportd
63152 node /Users/daniel/project/node_modules/.bin/next dev
`

	got := parsePSMetadata(output)
	if got[1107].CommandLine != "/usr/libexec/rapportd" {
		t.Fatalf("unexpected command line for 1107: %q", got[1107].CommandLine)
	}

	if got[1107].Path != "/usr/libexec/rapportd" {
		t.Fatalf("unexpected path for 1107: %q", got[1107].Path)
	}

	if got[63152].CommandLine != "node /Users/daniel/project/node_modules/.bin/next dev" {
		t.Fatalf("unexpected command line for 63152: %q", got[63152].CommandLine)
	}

	if got[63152].Path != "" {
		t.Fatalf("unexpected path fallback for 63152: %q", got[63152].Path)
	}
}
