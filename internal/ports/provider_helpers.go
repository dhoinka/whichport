package ports

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

type processMetadata struct {
	Path        string
	CommandLine string
}

func runCommand(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		return stdout.String(), nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return stdout.String(), nil
	}

	if stderr.Len() > 0 {
		return "", fmt.Errorf("%s failed: %s", name, strings.TrimSpace(stderr.String()))
	}

	return "", fmt.Errorf("%s failed: %w", name, err)
}

func uniquePIDs(listeners []partialListener) []int {
	seen := make(map[int]struct{}, len(listeners))
	var pids []int
	for _, listener := range listeners {
		if _, exists := seen[listener.PID]; exists {
			continue
		}

		seen[listener.PID] = struct{}{}
		pids = append(pids, listener.PID)
	}

	slices.Sort(pids)
	return pids
}

func chunkPIDs(pids []int, size int) [][]int {
	if size <= 0 || len(pids) == 0 {
		return nil
	}

	var chunks [][]int
	for start := 0; start < len(pids); start += size {
		end := start + size
		if end > len(pids) {
			end = len(pids)
		}
		chunks = append(chunks, pids[start:end])
	}

	return chunks
}

func joinPIDs(pids []int) string {
	parts := make([]string, 0, len(pids))
	for _, pid := range pids {
		parts = append(parts, strconv.Itoa(pid))
	}
	return strings.Join(parts, ",")
}
