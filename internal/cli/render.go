package cli

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/dhoinka/whichport/internal/ports"
	"golang.org/x/term"
)

type Renderer struct {
	out       io.Writer
	color     bool
	header    lipgloss.Style
	separator lipgloss.Style
	cell      lipgloss.Style
	accent    lipgloss.Style
	muted     lipgloss.Style
}

func NewRenderer(out io.Writer, color bool) Renderer {
	header := lipgloss.NewStyle().Bold(true)
	separator := lipgloss.NewStyle().Faint(true)
	cell := lipgloss.NewStyle()
	accent := lipgloss.NewStyle().Bold(true)
	muted := lipgloss.NewStyle().Faint(true)

	if color {
		header = header.Foreground(lipgloss.Color("86"))
		separator = separator.Foreground(lipgloss.Color("241"))
		accent = accent.Foreground(lipgloss.Color("219"))
		muted = muted.Foreground(lipgloss.Color("244"))
	}

	return Renderer{
		out:       out,
		color:     color,
		header:    header,
		separator: separator,
		cell:      cell,
		accent:    accent,
		muted:     muted,
	}
}

func (r Renderer) EmptyState() string {
	return r.muted.Render("No listening applications found.")
}

func (r Renderer) Render(listeners []ports.Listener) string {
	width := terminalWidth(r.out)
	portWidth := max(4, runeWidth("PORT"))
	protoWidth := max(5, runeWidth("PROTO"))
	pidWidth := runeWidth("PID")
	commandWidth := runeWidth("COMMAND")
	invocationWidth := runeWidth("INVOCATION")
	displayInvocations := make([]string, 0, len(listeners))

	for _, listener := range listeners {
		displayInvocation := normalizeInvocation(listener)
		displayInvocations = append(displayInvocations, displayInvocation)

		portWidth = max(portWidth, runeWidth(strconv.Itoa(listener.Port)))
		pidWidth = max(pidWidth, runeWidth(strconv.Itoa(listener.PID)))
		commandWidth = max(commandWidth, runeWidth(listener.Command))
		invocationWidth = max(invocationWidth, runeWidth(displayInvocation))
	}

	commandWidth = clamp(commandWidth, 12, 20)
	invocationWidth = clamp(invocationWidth, 28, 72)
	if width > 0 {
		remaining := width - (portWidth + protoWidth + pidWidth + commandWidth + 8)
		if remaining > 32 {
			invocationWidth = clamp(remaining, 32, 96)
		}
	}

	var lines []string
	headerLine := joinColumns(
		r.header.Render(padRight("PORT", portWidth)),
		r.header.Render(padRight("PROTO", protoWidth)),
		r.header.Render(padRight("PID", pidWidth)),
		r.header.Render(padRight("COMMAND", commandWidth)),
		r.header.Render("INVOCATION"),
	)
	lines = append(lines, headerLine)
	lines = append(lines, r.separator.Render(strings.Repeat("-", visibleWidth(headerLine))))

	for index, listener := range listeners {
		lines = append(lines, joinColumns(
			r.cell.Render(padRight(strconv.Itoa(listener.Port), portWidth)),
			r.accent.Render(padRight(strings.ToUpper(string(listener.Protocol)), protoWidth)),
			r.cell.Render(padRight(strconv.Itoa(listener.PID), pidWidth)),
			r.cell.Render(padRight(truncateMiddle(listener.Command, commandWidth), commandWidth)),
			r.cell.Render(truncateSuffix(displayInvocations[index], invocationWidth)),
		))
	}

	lines = append(lines, "")
	lines = append(lines, r.muted.Render(fmt.Sprintf("%d listener(s)", len(listeners))))
	return strings.Join(lines, "\n")
}

func isTerminal(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}

func terminalWidth(out io.Writer) int {
	file, ok := out.(*os.File)
	if !ok || !term.IsTerminal(int(file.Fd())) {
		return 0
	}

	width, _, err := term.GetSize(int(file.Fd()))
	if err != nil {
		return 0
	}

	return width
}

func joinColumns(cols ...string) string {
	return strings.Join(cols, "  ")
}

func padRight(value string, width int) string {
	padding := width - runeWidth(value)
	if padding <= 0 {
		return value
	}

	return value + strings.Repeat(" ", padding)
}

func truncateMiddle(value string, width int) string {
	if width <= 0 || runeWidth(value) <= width {
		return value
	}
	if width <= 3 {
		return strings.Repeat(".", width)
	}

	runes := []rune(value)
	remaining := width - 3
	head := remaining / 2
	tail := remaining - head
	return string(runes[:head]) + "..." + string(runes[len(runes)-tail:])
}

func truncateSuffix(value string, width int) string {
	if width <= 0 || runeWidth(value) <= width {
		return value
	}
	if width <= 3 {
		return strings.Repeat(".", width)
	}

	runes := []rune(value)
	return "..." + string(runes[len(runes)-(width-3):])
}

func normalizeInvocation(listener ports.Listener) string {
	command := strings.TrimSpace(listener.Command)
	commandLine := strings.TrimSpace(listener.CommandLine)
	path := strings.TrimSpace(listener.Path)

	if commandLine == "" {
		return command
	}

	if path != "" && path != "<unavailable>" && strings.HasPrefix(commandLine, path) {
		rest := strings.TrimSpace(strings.TrimPrefix(commandLine, path))
		if rest == "" {
			return command
		}
		if command == "" {
			return rest
		}
		return command + " " + rest
	}

	first, rest := splitFirstField(commandLine)
	switch {
	case rest == "" && command != "" && (first == command || strings.HasSuffix(first, "/"+command)):
		return command
	case rest != "" && command != "" && (first == command || strings.HasSuffix(first, "/"+command)):
		return command + " " + rest
	default:
		return commandLine
	}
}

func splitFirstField(value string) (string, string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ""
	}

	fields := strings.Fields(value)
	if len(fields) == 0 {
		return "", ""
	}
	if len(fields) == 1 {
		return fields[0], ""
	}

	return fields[0], strings.Join(fields[1:], " ")
}

func runeWidth(value string) int {
	return utf8.RuneCountInString(value)
}

func visibleWidth(value string) int {
	return lipgloss.Width(value)
}

func clamp(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
