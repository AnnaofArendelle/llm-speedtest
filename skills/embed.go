// Package skills embeds the AI skill package sources.
package skills

import "embed"

// FS contains skills/llm-speedtest.
//
//go:embed llm-speedtest
var FS embed.FS
