package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// prompter reads interactive answers from an input stream and writes prompts
// back to the output stream. It is intentionally minimal: it supports a single
// default value per prompt and validates against a caller-supplied list of
// allowed values when provided.
type prompter struct {
	reader *bufio.Reader
	writer io.Writer
}

func newPrompter(stdin io.Reader, stdout io.Writer) *prompter {
	return &prompter{reader: bufio.NewReader(stdin), writer: stdout}
}

// ask renders the prompt and returns the trimmed response. When the user
// provides an empty line the supplied default is returned instead.
func (p *prompter) ask(label, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Fprintf(p.writer, "%s [%s]: ", label, defaultValue)
	} else {
		fmt.Fprintf(p.writer, "%s: ", label)
	}

	value, err := p.readLine()
	if err != nil {
		return "", err
	}

	if value == "" {
		return defaultValue, nil
	}

	return value, nil
}

// askRequired prompts until the user supplies a non-empty value.
func (p *prompter) askRequired(label string) (string, error) {
	for {
		value, err := p.ask(label, "")
		if err != nil {
			return "", err
		}

		if value != "" {
			return value, nil
		}

		fmt.Fprintln(p.writer, "a value is required")
	}
}

// askChoice prompts the user for one of the supplied options. The default is
// accepted on an empty response.
func (p *prompter) askChoice(label string, options []string, defaultValue string) (string, error) {
	displayLabel := fmt.Sprintf("%s (%s)", label, strings.Join(options, "/"))

	for {
		value, err := p.ask(displayLabel, defaultValue)
		if err != nil {
			return "", err
		}

		if value == "" {
			return defaultValue, nil
		}

		for _, option := range options {
			if strings.EqualFold(value, option) {
				return option, nil
			}
		}

		fmt.Fprintf(p.writer, "must be one of: %s\n", strings.Join(options, ", "))
	}
}

// askBool returns true/false based on a y/n answer, defaulting to the supplied
// value on an empty response.
func (p *prompter) askBool(label string, defaultValue bool) (bool, error) {
	defaultLabel := "y/N"
	if defaultValue {
		defaultLabel = "Y/n"
	}

	displayLabel := fmt.Sprintf("%s (%s)", label, defaultLabel)

	for {
		value, err := p.ask(displayLabel, "")
		if err != nil {
			return false, err
		}

		if value == "" {
			return defaultValue, nil
		}

		switch strings.ToLower(value) {
		case "y", "yes", "true":
			return true, nil
		case "n", "no", "false":
			return false, nil
		}

		fmt.Fprintln(p.writer, "please answer y or n")
	}
}

// askList reads a comma-separated value and returns a trimmed, deduplicated
// slice. Empty input results in a nil slice.
func (p *prompter) askList(label string) ([]string, error) {
	value, err := p.ask(label, "")
	if err != nil {
		return nil, err
	}

	if value == "" {
		return nil, nil
	}

	parts := strings.Split(value, ",")
	seen := make(map[string]struct{}, len(parts))
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		result = append(result, trimmed)
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}

func (p *prompter) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil && (line == "" || err != io.EOF) {
		return "", err
	}

	return strings.TrimRight(line, "\r\n"), nil
}
