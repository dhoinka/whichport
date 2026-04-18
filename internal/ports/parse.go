package ports

import (
	"bufio"
	"bytes"
	"regexp"
	"strconv"
	"strings"
)

type partialListener struct {
	Port     int
	Protocol Protocol
	PID      int
	Command  string
}

var ssProcessPattern = regexp.MustCompile(`\("([^"]+)",pid=(\d+)`)

func parseSocketOutput(protocol Protocol, output string) []partialListener {
	var listeners []partialListener
	seen := map[string]struct{}{}

	var currentPID int
	var currentCommand string

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		switch line[0] {
		case 'p':
			currentPID, _ = strconv.Atoi(strings.TrimSpace(line[1:]))
			currentCommand = ""
		case 'c':
			currentCommand = strings.TrimSpace(line[1:])
		case 'n':
			port, ok := parsePort(strings.TrimSpace(line[1:]))
			if !ok || currentPID == 0 {
				continue
			}

			key := strings.Join([]string{
				string(protocol),
				strconv.Itoa(currentPID),
				strconv.Itoa(port),
			}, ":")
			if _, exists := seen[key]; exists {
				continue
			}

			listeners = append(listeners, partialListener{
				Port:     port,
				Protocol: protocol,
				PID:      currentPID,
				Command:  currentCommand,
			})
			seen[key] = struct{}{}
		}
	}

	return listeners
}

func parsePathOutput(output string) map[int]string {
	paths := make(map[int]string)

	var currentPID int
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		switch line[0] {
		case 'p':
			currentPID, _ = strconv.Atoi(strings.TrimSpace(line[1:]))
		case 'n':
			if currentPID == 0 || paths[currentPID] != "" {
				continue
			}

			path := strings.TrimSpace(line[1:])
			if strings.HasPrefix(path, "/") {
				paths[currentPID] = path
			}
		}
	}

	return paths
}

type psMetadata struct {
	Path        string
	CommandLine string
}

func parsePSMetadata(output string) map[int]psMetadata {
	result := make(map[int]psMetadata)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		splitAt := strings.IndexByte(line, ' ')
		if splitAt <= 0 || splitAt == len(line)-1 {
			continue
		}

		pid, err := strconv.Atoi(line[:splitAt])
		if err != nil {
			continue
		}

		commandLine := strings.TrimSpace(line[splitAt+1:])
		metadata := psMetadata{CommandLine: commandLine}
		if fields := strings.Fields(commandLine); len(fields) > 0 && strings.HasPrefix(fields[0], "/") {
			metadata.Path = fields[0]
		}

		result[pid] = metadata
	}

	return result
}

func parseSSOutput(protocol Protocol, output string) []partialListener {
	var listeners []partialListener
	seen := map[string]struct{}{}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		port, ok := parsePort(fields[4])
		if !ok {
			continue
		}

		for _, match := range ssProcessPattern.FindAllStringSubmatch(line, -1) {
			if len(match) != 3 {
				continue
			}

			pid, err := strconv.Atoi(match[2])
			if err != nil || pid == 0 {
				continue
			}

			key := strings.Join([]string{
				string(protocol),
				strconv.Itoa(pid),
				strconv.Itoa(port),
			}, ":")
			if _, exists := seen[key]; exists {
				continue
			}

			listeners = append(listeners, partialListener{
				Port:     port,
				Protocol: protocol,
				PID:      pid,
				Command:  match[1],
			})
			seen[key] = struct{}{}
		}
	}

	return listeners
}

func parseProcCmdline(output []byte) string {
	output = bytes.TrimRight(output, "\x00")
	if len(output) == 0 {
		return ""
	}

	parts := bytes.Split(output, []byte{0})
	args := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		args = append(args, string(part))
	}

	return strings.Join(args, " ")
}

func parsePort(name string) (int, bool) {
	if name == "" || strings.Contains(name, "->") {
		return 0, false
	}

	lastColon := strings.LastIndexByte(name, ':')
	if lastColon == -1 || lastColon == len(name)-1 {
		return 0, false
	}

	port, err := strconv.Atoi(name[lastColon+1:])
	if err != nil {
		return 0, false
	}

	return port, true
}
