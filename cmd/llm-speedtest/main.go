// Command llm-speedtest benchmarks OpenAI-compatible LLM APIs.
package main

import (
	"os"

	"github.com/AnnaofArendelle/llm-speedtest/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}
