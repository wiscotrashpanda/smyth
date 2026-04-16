package cli

import (
	"fmt"
	"io"
	"strings"
)

// bannerArt is the SMYTH wordmark rendered in the classic figlet "standard"
// font. Kept as pure ASCII so it looks right in every terminal, regardless of
// font support or color depth.
const bannerArt = ` ____  __  __ __   __ _____ _   _
/ ___||  \/  |\ \ / /|_   _| | | |
\___ \| |\/| | \ V /   | | | |_| |
 ___) | |  | |  | |    | | |  _  |
|____/|_|  |_|  |_|    |_| |_| |_|`

const bannerTagline = "forged manifests for anvil"

// writeBanner prints the wordmark and tagline to w, colored by s. It trails
// with a single blank line so callers can compose it with subsequent output
// without adding extra spacing.
func writeBanner(w io.Writer, s *styler) {
	lines := strings.Split(bannerArt, "\n")
	for _, line := range lines {
		fmt.Fprintln(w, s.forge(line))
	}

	// Indent the tagline to line up roughly under the wordmark.
	fmt.Fprintln(w)
	fmt.Fprintln(w, "      "+s.dim(bannerTagline))
	fmt.Fprintln(w)
}
