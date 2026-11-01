package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilter(t *testing.T) {
	f := DefaultFilter()
	if ok, _ := f.Match("x/combo-model", "combo"); ok {
		t.Fatal("combo must be excluded")
	}
	if ok, _ := f.Match("a/m", "a"); !ok {
		t.Fatal("default allows")
	}
	f = Filter{IncludeOwners: []string{"groq"}, IDRegex: "^groq/.*oss", Exclude: "guard"}
	if ok, _ := f.Match("groq/openai/gpt-oss-120b", "groq"); !ok {
		t.Fatal("include")
	}
	if ok, _ := f.Match("groq/gpt-oss-safeguard", "groq"); ok {
		t.Fatal("exclude regex")
	}
	if ok, _ := f.Match("cf/x", "cf"); ok {
		t.Fatal("owner filter")
	}
	if _, err := (Filter{IDRegex: "("}).Match("a", "b"); err == nil {
		t.Fatal("bad regex")
	}
	f = Filter{Models: []string{"a/b"}}
	if ok, _ := f.Match("a/b", "combo"); !ok {
		t.Fatal("explicit list wins")
	}
}

func TestSuiteAndFile(t *testing.T) {
	s := DefaultSuite(ProfileQuick)
	if s.Specs.Scan.Rounds() != 1 || s.LatencyTop != 5 || s.Specs.Quality.Trials != 1 {
		t.Fatalf("quick = %+v", s)
	}
	d := DefaultSuite(ProfileDeep)
	if d.Specs.Scan.Rounds() != 7 || len(d.Specs.Capability.LongContextTiers) != 3 {
		t.Fatalf("deep = %+v", d)
	}
	var z Suite
	z.Normalize()
	if z.LatencyTop == 0 || z.Lang != "zh" || z.Specs.Scan.First.Concurrency == 0 || len(z.Filter.ExcludeOwners) == 0 {
		t.Fatalf("normalize = %+v", z)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(path, []byte(Example()), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LLM_SPEEDTEST_KEY", "sk-env")
	f, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.Endpoint.BaseURL != "http://localhost:20128/v1" || f.Endpoint.APIKey != "sk-env" || f.Suite.Profile != "standard" || f.Suite.Specs.Scan.First.Concurrency != 10 || f.Suite.Specs.Load.SLO.TTFT == 0 {
		t.Fatalf("file = %+v", f)
	}
	if err := os.WriteFile(path, []byte("endpoint: [bad"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFile(path); err == nil {
		t.Fatal("bad yaml must fail")
	}
}
