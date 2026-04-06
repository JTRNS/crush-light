package chat

import "strings"

// looksLikeMarkdown checks if content appears to be markdown by looking for
// common markdown patterns.
func looksLikeMarkdown(content string) bool {
	patterns := []string{
		"# ",  // headers
		"## ", // headers
		"**",  // bold
		"```", // code fence
		"- ",  // unordered list
		"1. ", // ordered list
		"> ",  // blockquote
		"---", // horizontal rule
		"***", // horizontal rule
	}
	for _, p := range patterns {
		if strings.Contains(content, p) {
			return true
		}
	}
	return false
}
