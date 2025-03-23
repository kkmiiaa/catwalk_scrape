package utils

import "strings"

// https://zenn.dev/mkosakana/articles/54eff8a2039084

func TrimString(
	text string,
	prefix string,
	suffix string,
) string {
	text = strings.TrimLeft(text, prefix)
	text = strings.TrimRight(text, suffix)
	text = strings.TrimSpace(text)

	return text
}
