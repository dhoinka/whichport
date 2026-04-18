//go:build linux

package ports

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func discoverPlatform(ctx context.Context, query Query) ([]Listener, error) {
	var partials []partialListener
	for _, protocol := range query.Protocols {
		args := []string{"-H", "-O", "-n", "-l", "-p"}
		switch protocol {
		case ProtocolTCP:
			args = append(args, "-t")
		case ProtocolUDP:
			args = append(args, "-u")
		default:
			return nil, fmt.Errorf("unsupported protocol %q", protocol)
		}

		output, err := runCommand(ctx, "ss", args...)
		if err != nil {
			return nil, err
		}

		partials = append(partials, filterPartialsByPort(parseSSOutput(protocol, output), query.Port)...)
	}

	metadata, err := lookupLinuxProcessMetadata(uniquePIDs(partials))
	if err != nil {
		return nil, err
	}

	listeners := make([]Listener, 0, len(partials))
	for _, partial := range partials {
		details := metadata[partial.PID]

		path := strings.TrimSpace(details.Path)
		if path == "" {
			path = "<unavailable>"
		}

		commandLine := strings.TrimSpace(details.CommandLine)
		if commandLine == "" {
			commandLine = partial.Command
		}

		listeners = append(listeners, Listener{
			Port:        partial.Port,
			Protocol:    partial.Protocol,
			PID:         partial.PID,
			Command:     partial.Command,
			CommandLine: commandLine,
			Path:        path,
		})
	}

	return listeners, nil
}

func filterPartialsByPort(partials []partialListener, port int) []partialListener {
	if port == 0 {
		return partials
	}

	filtered := make([]partialListener, 0, len(partials))
	for _, partial := range partials {
		if partial.Port == port {
			filtered = append(filtered, partial)
		}
	}

	return filtered
}

func lookupLinuxProcessMetadata(pids []int) (map[int]processMetadata, error) {
	metadata := make(map[int]processMetadata, len(pids))
	for _, pid := range pids {
		procRoot := filepath.Join("/proc", strconv.Itoa(pid))
		details := processMetadata{}

		path, err := os.Readlink(filepath.Join(procRoot, "exe"))
		switch {
		case err == nil:
			details.Path = path
		case isIgnorableProcError(err):
		default:
			return nil, err
		}

		cmdlineBytes, err := os.ReadFile(filepath.Join(procRoot, "cmdline"))
		switch {
		case err == nil:
			details.CommandLine = parseProcCmdline(cmdlineBytes)
		case isIgnorableProcError(err):
		default:
			return nil, err
		}

		metadata[pid] = details
	}

	return metadata, nil
}

func isIgnorableProcError(err error) bool {
	return err == nil || errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrPermission)
}
