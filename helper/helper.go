package helper

import "strings"

func EscapeTelegramReserved(s string) string {

	s = strings.ReplaceAll(s, `\`, `\\`)

	s = strings.ReplaceAll(s, "(", "\\(")

	s = strings.ReplaceAll(s, ")", "\\)")

	s = strings.ReplaceAll(s, ".", "\\.")

	s = strings.ReplaceAll(s, "_", "\\_")

	s = strings.ReplaceAll(s, "|", "\\|")

	s = strings.ReplaceAll(s, "-", "\\-")

	return s
}
