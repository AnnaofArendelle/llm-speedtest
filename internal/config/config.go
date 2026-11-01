// Package config defines endpoint settings, run specifications and the
// suite profiles, plus YAML loading for the CLI.
package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/AnnaofArendelle/llm-speedtest/internal/analyze"
	"github.com/AnnaofArendelle/llm-speedtest/internal/capability"
	"github.com/AnnaofArendelle/llm-speedtest/internal/latency"
	"github.com/AnnaofArendelle/llm-speedtest/internal/load"
	"github.com/AnnaofArendelle/llm-speedtest/internal/quality"
	"github.com/AnnaofArendelle/llm-speedtest/internal/scan"
)

// Endpoint describes where to send requests. The API key is never persisted
// by this package.
type Endpoint struct {
	Name     string            `json:"name" yaml:"name"`
	BaseURL  string            `json:"base_url" yaml:"base_url"`
	APIKey   string            `json:"-" yaml:"api_key"`
	Dialect  string            `json:"dialect" yaml:"dialect"` // empty = auto-detect
	Headers  map[string]string `json:"headers,omitempty" yaml:"headers"`
	Insecure bool              `json:"insecure" yaml:"insecure"`
	Proxy    string            `json:"proxy,omitempty" yaml:"proxy"`
}

// Filter selects which discovered models take part.
type Filter struct {
	ExcludeOwners []string `json:"exclude_owners" yaml:"exclude_owners"` // default: ["combo"]
	IncludeOwners []string `json:"include_owners" yaml:"include_owners"`
	IDRegex       string   `json:"id_regex" yaml:"id_regex"`
	Exclude       string   `json:"exclude_regex" yaml:"exclude_regex"`
	Limit         int      `json:"limit" yaml:"limit"`
	Models        []string `json:"models" yaml:"models"` // explicit list (skips discovery filtering)
}

// DefaultFilter excludes aggregated pseudo-models.
func DefaultFilter() Filter { return Filter{ExcludeOwners: []string{"combo"}} }

// Match reports whether a model passes the filter.
func (f Filter) Match(id, ownedBy string) (bool, error) {
	if len(f.Models) > 0 {
		for _, m := range f.Models {
			if m == id {
				return true, nil
			}
		}
		return false, nil
	}
	for _, o := range f.ExcludeOwners {
		if strings.EqualFold(o, ownedBy) {
			return false, nil
		}
	}
	if len(f.IncludeOwners) > 0 {
		ok := false
		for _, o := range f.IncludeOwners {
			if strings.EqualFold(o, ownedBy) {
				ok = true
			}
		}
		if !ok {
			return false, nil
		}
	}
	if f.IDRegex != "" {
		re, err := regexp.Compile(f.IDRegex)
		if err != nil {
			return false, fmt.Errorf("id_regex: %w", err)
		}
		if !re.MatchString(id) {
			return false, nil
		}
	}
	if f.Exclude != "" {
		re, err := regexp.Compile(f.Exclude)
		if err != nil {
			return false, fmt.Errorf("exclude_regex: %w", err)
		}
		if re.MatchString(id) {
			return false, nil
		}
	}
	return true, nil
}

// Specs bundles every engine specification.
type Specs struct {
	Scan       scan.Spec       `json:"scan" yaml:"scan"`
	Latency    latency.Spec    `json:"latency" yaml:"latency"`
	Load       load.Spec       `json:"load" yaml:"load"`
	Capability capability.Spec `json:"capability" yaml:"capability"`
	Quality    quality.Spec    `json:"quality" yaml:"quality"`
	Scoring    analyze.Scoring `json:"scoring" yaml:"scoring"`
}

// Normalize fills zero values in every spec.
func (s *Specs) Normalize() {
	s.Scan.Normalize()
	s.Latency.Normalize()
	s.Load.Normalize()
	s.Capability.Normalize()
	s.Quality.Normalize()
	s.Scoring.Normalize()
}

// Suite describes a full pipeline run.
type Suite struct {
	Profile       string `json:"profile" yaml:"profile"` // quick | standard | deep
	Scan          bool   `json:"scan" yaml:"scan"`
	Latency       bool   `json:"latency" yaml:"latency"`
	LatencyTop    int    `json:"latency_top" yaml:"latency_top"` // models taken from the scan ranking
	Capability    bool   `json:"capability" yaml:"capability"`
	CapabilityTop int    `json:"capability_top" yaml:"capability_top"`
	Quality       bool   `json:"quality" yaml:"quality"`
	QualityTop    int    `json:"quality_top" yaml:"quality_top"`
	Load          bool   `json:"load" yaml:"load"`
	LoadModel     string `json:"load_model" yaml:"load_model"` // empty = best model from the scan
	Filter        Filter `json:"filter" yaml:"filter"`
	Specs         Specs  `json:"specs" yaml:"specs"`
	Lang          string `json:"lang" yaml:"lang"`
}

// Profiles.
const (
	ProfileQuick    = "quick"
	ProfileStandard = "standard"
	ProfileDeep     = "deep"
)

// DefaultSuite returns the suite for a profile.
func DefaultSuite(profile string) Suite {
	if profile == "" {
		profile = ProfileStandard
	}
	s := Suite{Profile: profile, Scan: true, Latency: true, LatencyTop: 10, Capability: true, CapabilityTop: 10, Quality: true, QualityTop: 10, Filter: DefaultFilter(), Lang: "zh"}
	s.Specs.Scan = scan.Preset(profile)
	s.Specs.Latency = latency.Default()
	s.Specs.Load = load.Default()
	s.Specs.Capability = capability.Default()
	s.Specs.Quality = quality.DefaultSpec()
	s.Specs.Scoring = analyze.DefaultScoring()
	switch profile {
	case ProfileQuick:
		s.LatencyTop, s.CapabilityTop, s.QualityTop = 5, 5, 5
		s.Specs.Latency.Samples, s.Specs.Latency.Warmup = 3, 1
		s.Specs.Quality.Trials = 1
		s.Specs.Capability.LongContextTiers = []int{8000}
	case ProfileDeep:
		s.LatencyTop, s.CapabilityTop, s.QualityTop = 20, 20, 20
		s.Specs.Latency.Samples = 8
		s.Specs.Quality.Trials = 3
		s.Specs.Capability.LongContextTiers = []int{8000, 32000, 128000}
	}
	return s
}

// Normalize fills defaults.
func (s *Suite) Normalize() {
	d := DefaultSuite(s.Profile)
	if s.LatencyTop <= 0 {
		s.LatencyTop = d.LatencyTop
	}
	if s.CapabilityTop <= 0 {
		s.CapabilityTop = d.CapabilityTop
	}
	if s.QualityTop <= 0 {
		s.QualityTop = d.QualityTop
	}
	if s.Lang == "" {
		s.Lang = "zh"
	}
	if len(s.Filter.ExcludeOwners) == 0 && len(s.Filter.Models) == 0 && len(s.Filter.IncludeOwners) == 0 && s.Filter.IDRegex == "" {
		s.Filter.ExcludeOwners = []string{"combo"}
	}
	if s.Specs.Scan.First.Concurrency == 0 {
		s.Specs.Scan = scan.Preset(s.Profile)
	}
	s.Specs.Normalize()
}

// File is the YAML configuration accepted by the CLI.
type File struct {
	Endpoint Endpoint `yaml:"endpoint"`
	Suite    Suite    `yaml:"suite"`
}

// LoadFile reads a YAML config.
func LoadFile(path string) (*File, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f := &File{Suite: DefaultSuite(ProfileStandard)}
	if err := yaml.Unmarshal(raw, f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	f.Suite.Normalize()
	if f.Endpoint.APIKey == "" {
		f.Endpoint.APIKey = os.Getenv("LLM_SPEEDTEST_KEY")
	}
	return f, nil
}

// Example returns an annotated example YAML.
func Example() string {
	return `# llm-speedtest configuration
endpoint:
  name: my-gateway
  base_url: http://localhost:20128/v1
  api_key: ""            # or set LLM_SPEEDTEST_KEY
  dialect: ""            # auto-detect; or 9router | oneapi | openrouter | litellm | ollama | vllm | generic
  headers: {}
  insecure: false
  proxy: ""
suite:
  profile: standard      # quick | standard | deep
  scan: true
  latency: true
  latency_top: 10
  capability: true
  capability_top: 10
  quality: true
  quality_top: 10
  load: false
  load_model: ""
  lang: zh
  filter:
    exclude_owners: [combo]
    include_owners: []
    id_regex: ""
    exclude_regex: ""
    limit: 0
  specs:
    scan:
      first: {concurrency: 10, per_provider: 2, timeout: 60s}
      retry: {concurrency: 6, per_provider: 1, timeout: 90s}
      final: {concurrency: 4, per_provider: 1, timeout: 90s}
      resample: 2
      max_tokens: 1
      prompt: hi
      max_cooldown_wait: 120s
    latency: {samples: 5, warmup: 1, input_tokens: 200, output_tokens: 128, concurrency: 2, per_provider: 1, timeout: 120s}
    load: {mode: concurrency, levels: [1, 2, 4], duration: 30s, input_tokens: 100, output_tokens: 64, stream: true, slo: {ttft: 1s, tpot: 50ms}}
    capability: {concurrency: 2, per_provider: 1, timeout: 90s, long_context_tiers: [8000, 32000]}
    quality: {trials: 3, concurrency: 2, per_provider: 1, timeout: 60s, max_tokens: 64}
    scoring: {w_stability: 0.4, w_breadth: 0.3, w_speed: 0.3, speed_lo_s: 0.5, speed_hi_s: 30}
`
}
