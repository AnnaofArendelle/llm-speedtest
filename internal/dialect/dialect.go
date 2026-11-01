// Package dialect detects which kind of gateway/server sits behind an
// OpenAI-compatible endpoint and exposes the small, structural extras each
// one offers (health/version probes). Dialects never carry naming tables.
package dialect

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/apiclient"
)

// Info describes a detected dialect.
type Info struct {
	Name     string   `json:"name"`
	Version  string   `json:"version,omitempty"`
	Latest   string   `json:"latest,omitempty"`
	Healthy  *bool    `json:"healthy,omitempty"`
	Evidence []string `json:"evidence,omitempty"`
	Score    int      `json:"score"`
}

// Dialect probes an endpoint for its fingerprint.
type Dialect interface {
	Name() string
	Detect(ctx context.Context, c *apiclient.Client) Info
}

// Names of the built-in dialects.
const (
	Generic    = "generic"
	NineRouter = "9router"
	OneAPI     = "oneapi"
	OpenRouter = "openrouter"
	LiteLLM    = "litellm"
	Ollama     = "ollama"
	VLLM       = "vllm"
)

// All returns the built-in dialects.
func All() []Dialect {
	return []Dialect{nineRouter{}, oneAPI{}, openRouter{}, liteLLM{}, ollama{}, vllm{}}
}

// Names lists selectable dialect names.
func Names() []string {
	return []string{Generic, NineRouter, OneAPI, OpenRouter, LiteLLM, Ollama, VLLM}
}

// Detect runs every dialect probe concurrently and returns the best match;
// "generic" when nothing scores.
func Detect(ctx context.Context, c *apiclient.Client) Info {
	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	results := make([]Info, len(All()))
	var wg sync.WaitGroup
	for i, d := range All() {
		wg.Add(1)
		go func(i int, d Dialect) {
			defer wg.Done()
			results[i] = d.Detect(ctx, c)
			results[i].Name = d.Name()
		}(i, d)
	}
	wg.Wait()
	best := Info{Name: Generic}
	for _, r := range results {
		if r.Score > best.Score {
			best = r
		}
	}
	return best
}

// Health re-runs the probe of a named dialect to obtain fresh health/version
// information. Unknown or generic dialects return an empty Info.
func Health(ctx context.Context, name string, c *apiclient.Client) Info {
	for _, d := range All() {
		if d.Name() == name {
			ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
			defer cancel()
			info := d.Detect(ctx, c)
			info.Name = name
			return info
		}
	}
	return Info{Name: Generic}
}

func getJSON(ctx context.Context, c *apiclient.Client, u string, limit int64) (int, map[string]any, []byte) {
	status, body, _, err := c.Get(ctx, u, limit)
	if err != nil {
		return 0, nil, nil
	}
	var m map[string]any
	_ = json.Unmarshal(body, &m)
	return status, m, body
}

type nineRouter struct{}

func (nineRouter) Name() string { return NineRouter }
func (nineRouter) Detect(ctx context.Context, c *apiclient.Client) Info {
	info := Info{}
	status, m, _ := getJSON(ctx, c, c.Origin()+"/api/version", 64*1024)
	if status == 200 && m != nil {
		if v, ok := m["currentVersion"].(string); ok {
			info.Score = 100
			info.Version = v
			info.Latest, _ = m["latestVersion"].(string)
			info.Evidence = append(info.Evidence, "/api/version reports currentVersion")
		}
	}
	if info.Score > 0 {
		hs, _, _ := getJSON(ctx, c, c.Origin()+"/api/health", 4096)
		ok := hs == 200
		info.Healthy = &ok
	}
	return info
}

type oneAPI struct{}

func (oneAPI) Name() string { return OneAPI }
func (oneAPI) Detect(ctx context.Context, c *apiclient.Client) Info {
	info := Info{}
	status, m, _ := getJSON(ctx, c, c.Origin()+"/api/status", 256*1024)
	if status == 200 && m != nil {
		if data, ok := m["data"].(map[string]any); ok {
			if v, ok := data["version"].(string); ok {
				info.Score = 90
				info.Version = v
				info.Evidence = append(info.Evidence, "/api/status reports data.version")
				if name, ok := data["system_name"].(string); ok && name != "" {
					info.Evidence = append(info.Evidence, "system_name="+name)
				}
				ok := true
				info.Healthy = &ok
			}
		}
	}
	return info
}

type openRouter struct{}

func (openRouter) Name() string { return OpenRouter }
func (openRouter) Detect(_ context.Context, c *apiclient.Client) Info {
	if strings.Contains(strings.ToLower(c.Origin()), "openrouter.ai") {
		return Info{Score: 100, Evidence: []string{"host is openrouter.ai"}}
	}
	return Info{}
}

type liteLLM struct{}

func (liteLLM) Name() string { return LiteLLM }
func (liteLLM) Detect(ctx context.Context, c *apiclient.Client) Info {
	info := Info{}
	status, body, hdr, err := c.Get(ctx, c.Origin()+"/health/liveliness", 4096)
	if err == nil && status == 200 && strings.Contains(strings.ToLower(string(body)), "alive") {
		info.Score = 95
		info.Evidence = append(info.Evidence, "/health/liveliness")
		ok := true
		info.Healthy = &ok
	}
	if v := hdr.Get("x-litellm-version"); v != "" {
		info.Score = 100
		info.Version = v
		info.Evidence = append(info.Evidence, "x-litellm-version header")
	}
	return info
}

type ollama struct{}

func (ollama) Name() string { return Ollama }
func (ollama) Detect(ctx context.Context, c *apiclient.Client) Info {
	info := Info{}
	status, m, _ := getJSON(ctx, c, c.Origin()+"/api/tags", 4*1024*1024)
	if status == 200 && m != nil {
		if _, ok := m["models"]; ok {
			info.Score = 95
			info.Evidence = append(info.Evidence, "/api/tags lists models")
			ok := true
			info.Healthy = &ok
		}
	}
	if vs, vm, _ := getJSON(ctx, c, c.Origin()+"/api/version", 4096); vs == 200 && vm != nil {
		if v, ok := vm["version"].(string); ok && info.Score > 0 {
			info.Version = v
		}
	}
	return info
}

type vllm struct{}

func (vllm) Name() string { return VLLM }
func (vllm) Detect(ctx context.Context, c *apiclient.Client) Info {
	info := Info{}
	status, m, _ := getJSON(ctx, c, c.Origin()+"/version", 4096)
	if status == 200 && m != nil {
		if v, ok := m["version"].(string); ok {
			info.Score = 80
			info.Version = v
			info.Evidence = append(info.Evidence, "/version reports version")
			ok := true
			info.Healthy = &ok
		}
	}
	return info
}
