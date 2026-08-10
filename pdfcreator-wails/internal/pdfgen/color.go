package pdfgen

import (
	"fmt"
	"strconv"
	"strings"
)

// parseHexColor converts #RRGGBB or #RGB to RGB int values.
func parseHexColor(hex string) (int, int, int, error) {
	hex = strings.TrimSpace(hex)
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}

	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %s", hex)
	}

	v, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}

	r := int((v >> 16) & 0xFF)
	g := int((v >> 8) & 0xFF)
	b := int(v & 0xFF)

	return r, g, b, nil
}
