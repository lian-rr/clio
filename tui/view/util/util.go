package util

import (
	"bytes"

	"github.com/alecthomas/chroma/v2/quick"
)

const (
	chromaLang      = "fish"
	chromaFormatter = "terminal16m"
	chromaStyle     = "catppuccin-frappe"
)

// RelativeDimensions returns the dimensions based on the desired percentages.
func RelativeDimensions(w, h int, pw, ph float32) (width, height int) {
	return int(float32(w) * pw), int(float32(h) * ph)
}

// FormatCommand returns the passed command formatted
func FormatCommand(raw string) (string, error) {
	var b bytes.Buffer
	if err := quick.Highlight(&b, raw, chromaLang, chromaFormatter, chromaStyle); err != nil {
		return "", err
	}

	return b.String(), nil
}
