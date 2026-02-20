// Package iterm provides iTerm2-specific terminal escape sequences.
package iterm

import (
	"fmt"
	"hash/fnv"
	"math"
	"os"
)

// SetTabTitle sets the terminal tab title. Works in iTerm2 and most terminals
// that support xterm title escape sequences.
func SetTabTitle(title string) {
	//nolint:errcheck // best-effort terminal escape sequence
	fmt.Fprintf(os.Stdout, "\033]0;%s\a", title)
}

// SetTabColor sets the iTerm2 tab color based on a hash of the given key.
// Each unique key gets a consistent, distinct color. No-op outside iTerm2.
func SetTabColor(key string) {
	if os.Getenv("TERM_PROGRAM") != "iTerm.app" {
		return
	}
	r, g, b := keyToRGB(key)
	//nolint:errcheck // best-effort terminal escape sequences
	fmt.Fprintf(os.Stdout, "\033]6;1;bg;red;brightness;%d\a", r)
	//nolint:errcheck
	fmt.Fprintf(os.Stdout, "\033]6;1;bg;green;brightness;%d\a", g)
	//nolint:errcheck
	fmt.Fprintf(os.Stdout, "\033]6;1;bg;blue;brightness;%d\a", b)
}

// keyToRGB hashes a key to a hue and converts to RGB with fixed saturation
// and lightness for visually pleasant, distinct colors.
func keyToRGB(key string) (uint8, uint8, uint8) {
	h := fnv.New32a()
	h.Write([]byte(key))
	hue := float64(h.Sum32()%360) / 360.0
	return hslToRGB(hue, 0.6, 0.45)
}

func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h*6, 2)-1))
	m := l - c/2
	var r, g, b float64
	switch {
	case h < 1.0/6:
		r, g, b = c, x, 0
	case h < 2.0/6:
		r, g, b = x, c, 0
	case h < 3.0/6:
		r, g, b = 0, c, x
	case h < 4.0/6:
		r, g, b = 0, x, c
	case h < 5.0/6:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return uint8((r + m) * 255), uint8((g + m) * 255), uint8((b + m) * 255)
}
