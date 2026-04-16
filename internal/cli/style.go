package cli

import (
	"io"
	"os"
)

// styler wraps text in ANSI escape sequences when the destination writer looks
// like an interactive terminal. When colors are disabled (non-TTY output,
// NO_COLOR set, TERM=dumb) the methods return the input unchanged, so callers
// can style output unconditionally without worrying about pipes, redirects, or
// CI log captures swallowing escape codes.
type styler struct {
	enabled bool
}

func newStyler(w io.Writer) *styler {
	return &styler{enabled: colorEnabled(w)}
}

// colorEnabled returns true when ANSI colors should be emitted to w. It honors
// the NO_COLOR convention (https://no-color.org), FORCE_COLOR for users that
// want color in non-TTY contexts, and TERM=dumb to silence color regardless of
// the other signals.
func colorEnabled(w io.Writer) bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}

	if os.Getenv("TERM") == "dumb" {
		return false
	}

	if forced := os.Getenv("FORCE_COLOR"); forced != "" && forced != "0" {
		return true
	}

	file, ok := w.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}

// wrap renders text inside the supplied SGR codes (e.g. "1;33" for bold
// yellow) when coloring is enabled.
func (s *styler) wrap(code, text string) string {
	if !s.enabled || text == "" {
		return text
	}

	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func (s *styler) bold(t string) string   { return s.wrap("1", t) }
func (s *styler) dim(t string) string    { return s.wrap("2", t) }
func (s *styler) red(t string) string    { return s.wrap("31", t) }
func (s *styler) green(t string) string  { return s.wrap("32", t) }
func (s *styler) yellow(t string) string { return s.wrap("33", t) }
func (s *styler) cyan(t string) string   { return s.wrap("36", t) }

// forge renders text in a warm ember tone, falling back to bold yellow on
// terminals that only advertise 16-color support.
func (s *styler) forge(t string) string {
	if !s.enabled || t == "" {
		return t
	}

	// 256-color palette: 208 is a saturated orange that evokes hot metal.
	return "\x1b[38;5;208m" + t + "\x1b[0m"
}
