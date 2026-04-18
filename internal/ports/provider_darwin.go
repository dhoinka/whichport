//go:build darwin

package ports

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

func discoverPlatform(ctx context.Context, query Query) ([]Listener, error) {
	var partials []partialListener
	for _, protocol := range query.Protocols {
		args := []string{"-nP", "-Fpcn"}
		switch protocol {
		case ProtocolTCP:
			if query.Port > 0 {
				args = append(args, fmt.Sprintf("-iTCP:%d", query.Port))
			} else {
				args = append(args, "-iTCP")
			}
			args = append(args, "-sTCP:LISTEN")
		case ProtocolUDP:
			if query.Port > 0 {
				args = append(args, fmt.Sprintf("-iUDP:%d", query.Port))
			} else {
				args = append(args, "-iUDP")
			}
		default:
			return nil, fmt.Errorf("unsupported protocol %q", protocol)
		}

		output, err := runCommand(ctx, "lsof", args...)
		if err != nil {
			return nil, err
		}

		partials = append(partials, parseSocketOutput(protocol, output)...)
	}

	pids := uniquePIDs(partials)
	metadata, err := lookupProcessMetadata(ctx, pids)
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

func lookupProcessMetadata(ctx context.Context, pids []int) (map[int]processMetadata, error) {
	if len(pids) == 0 {
		return map[int]processMetadata{}, nil
	}

	paths, err := lookupPaths(ctx, pids)
	if err != nil {
		return nil, err
	}

	psMetadata, err := lookupPSMetadata(ctx, pids)
	if err != nil {
		return nil, err
	}

	metadata := make(map[int]processMetadata, len(pids))
	for _, pid := range pids {
		details := processMetadata{
			Path:        paths[pid],
			CommandLine: psMetadata[pid].CommandLine,
		}
		if details.Path == "" {
			details.Path = psMetadata[pid].Path
		}
		metadata[pid] = details
	}

	return metadata, nil
}

func lookupPaths(ctx context.Context, pids []int) (map[int]string, error) {
	paths := make(map[int]string, len(pids))
	for _, chunk := range chunkPIDs(pids, 128) {
		args := []string{"-Fn", "-a", "-d", "txt", "-p", joinPIDs(chunk)}
		output, err := runCommand(ctx, "lsof", args...)
		if err != nil {
			return nil, err
		}

		for pid, path := range parsePathOutput(output) {
			paths[pid] = path
		}
	}

	return paths, nil
}

func lookupPSMetadata(ctx context.Context, pids []int) (map[int]psMetadata, error) {
	args := []string{"-ww", "-o", "pid=", "-o", "command="}
	for _, pid := range pids {
		args = append(args, "-p", strconv.Itoa(pid))
	}

	output, err := runCommand(ctx, "ps", args...)
	if err != nil {
		return nil, err
	}

	return parsePSMetadata(output), nil
}
