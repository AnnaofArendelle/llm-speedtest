package modelid

import "testing"

func TestClean(t *testing.T) {
	cases := []struct {
		id, owner, name, org, prefix string
		stripped                     bool
	}{
		{"kilo/deepseek/deepseek-v4-flash", "kilo", "deepseek-v4-flash", "deepseek", "kilo", true},
		{"RouteWay/deepseek-v4.1-flash:free", "RouteWay", "deepseek-v4.1-flash:free", "", "RouteWay", true},
		{"routeway/x", "RouteWay", "x", "", "routeway", true},
		{"cf/@cf/meta/llama-3.1-8b-instruct", "cf", "llama-3.1-8b-instruct", "cf-meta", "cf", true},
		{"openai/gpt-4o", "", "gpt-4o", "openai", "openai", false},
		{"gpt-4o", "openai", "gpt-4o", "", "", false},
		{"kilo/~openai/gpt-5[1m]", "kilo", "gpt-5", "openai", "kilo", true},
		{"kilo/", "kilo", "kilo", "", "kilo", false},
		{"", "", "", "", "", false},
	}
	for _, c := range cases {
		p := Clean(c.id, c.owner)
		if p.Name != c.name || p.Org != c.org || p.Prefix != c.prefix || p.OwnerStripped != c.stripped {
			t.Errorf("Clean(%q,%q) = %+v, want name=%q org=%q prefix=%q stripped=%v", c.id, c.owner, p, c.name, c.org, c.prefix, c.stripped)
		}
	}
	if Upstream("kios/ling-3.1-flash", "kios") != "ling-3.1-flash" || Upstream("x", "kios") != "x" {
		t.Error("Upstream mismatch")
	}
}
