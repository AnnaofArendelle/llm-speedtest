// Package provider decides, per endpoint, which field of the discovered
// model list identifies the provider ("channel") a model is served through.
// Nothing is hard-coded: the decision is made from the discriminating power
// of each candidate field and its consistency with connection identifiers
// observed in gateway error messages, and users may override the result.
package provider

import (
	"sort"
	"strconv"
	"strings"

	"github.com/AnnaofArendelle/llm-speedtest/internal/modelid"
)

// Model is the discovery record the resolver works on.
type Model struct {
	ID      string
	OwnedBy string
}

// Evidence records one observation that supports an assignment.
type Evidence struct {
	Field string `json:"field"` // owned_by | prefix | conn | endpoint
	Value string `json:"value"`
}

// Assignment is the provider of one model.
type Assignment struct {
	Key      string     `json:"key"`  // stable key used for grouping and concurrency limits
	Name     string     `json:"name"` // display name
	Evidence []Evidence `json:"evidence"`
}

// Override renames a raw provider key or merges it into another one.
type Override struct {
	DisplayName string
	MergedInto  string
}

// Suggestion proposes merging raw keys that share a connection identifier.
type Suggestion struct {
	Keys     []string `json:"keys"`
	Evidence string   `json:"evidence"`
}

// Result is the full resolution for an endpoint.
type Result struct {
	Field       string                `json:"field"` // field chosen as the provider key
	Rationale   string                `json:"rationale"`
	Models      map[string]Assignment `json:"models"`
	Providers   []Provider            `json:"providers"`
	Suggestions []Suggestion          `json:"suggestions"`
}

// Provider summarises one resolved provider.
type Provider struct {
	Key    string   `json:"key"`
	Name   string   `json:"name"`
	Models int      `json:"models"`
	Conns  []string `json:"conns,omitempty"`
	Raw    []string `json:"raw_keys,omitempty"` // raw keys merged into this provider
}

// generic owned_by values carry no channel information on their own.
var genericOwners = map[string]bool{
	"": true, "system": true, "library": true, "user": true, "organization-owner": true, "custom": true,
	"openai-compatible": true, "default": true, "unknown": true, "model": true, "null": true, "none": true, "n/a": true,
}

// Resolve chooses the provider field for the endpoint and assigns every
// model. conns maps a model id to the connection identifier extracted from
// gateway error messages (may be empty). endpointName labels the single
// provider case.
func Resolve(models []Model, conns map[string]string, overrides map[string]Override, endpointName string) Result {
	res := Result{Models: make(map[string]Assignment, len(models))}
	n := len(models)
	ownerVals := map[string]int{}
	prefixVals := map[string]int{}
	for _, m := range models {
		o := strings.TrimSpace(m.OwnedBy)
		if !genericOwners[strings.ToLower(o)] {
			ownerVals[o]++
		}
		if p := modelid.Clean(m.ID, "").Prefix; p != "" {
			prefixVals[p]++
		}
	}
	ownerConsistency, ownerCovered := consistency(models, conns, func(m Model) string { return strings.TrimSpace(m.OwnedBy) })
	prefixConsistency, prefixCovered := consistency(models, conns, func(m Model) string { return modelid.Clean(m.ID, "").Prefix })

	switch {
	case len(ownerVals) >= 2 && coverage(ownerVals, n) >= 0.5 && ownerConsistency >= 0.9:
		res.Field = "owned_by"
		res.Rationale = rationale("owned_by", len(ownerVals), ownerCovered, ownerConsistency)
	case len(prefixVals) >= 2 && coverage(prefixVals, n) >= 0.5 && prefixConsistency >= 0.9:
		res.Field = "prefix"
		res.Rationale = rationale("id prefix", len(prefixVals), prefixCovered, prefixConsistency)
	case len(ownerVals) >= 2 && coverage(ownerVals, n) >= 0.5:
		res.Field = "owned_by"
		res.Rationale = rationale("owned_by", len(ownerVals), ownerCovered, ownerConsistency) + "; connection identifiers disagree, review suggestions"
	default:
		res.Field = "endpoint"
		res.Rationale = "no field discriminates between models; treating the endpoint as a single provider"
	}

	rawKey := func(m Model) string {
		switch res.Field {
		case "owned_by":
			if o := strings.TrimSpace(m.OwnedBy); !genericOwners[strings.ToLower(o)] {
				return o
			}
			if p := modelid.Clean(m.ID, "").Prefix; p != "" {
				return p
			}
			return endpointLabel(endpointName)
		case "prefix":
			if p := modelid.Clean(m.ID, "").Prefix; p != "" {
				return p
			}
			if o := strings.TrimSpace(m.OwnedBy); !genericOwners[strings.ToLower(o)] {
				return o
			}
			return endpointLabel(endpointName)
		}
		return endpointLabel(endpointName)
	}

	provs := map[string]*Provider{}
	connsByKey := map[string]map[string]int{}
	for _, m := range models {
		raw := rawKey(m)
		key, name := raw, raw
		if ov, ok := overrides[raw]; ok {
			if ov.MergedInto != "" {
				key = ov.MergedInto
				if ov2, ok := overrides[key]; ok && ov2.DisplayName != "" {
					name = ov2.DisplayName
				} else {
					name = key
				}
			}
			if ov.DisplayName != "" && ov.MergedInto == "" {
				name = ov.DisplayName
			}
		}
		ev := []Evidence{{Field: res.Field, Value: raw}}
		if c := conns[m.ID]; c != "" {
			ev = append(ev, Evidence{Field: "conn", Value: c})
			if connsByKey[key] == nil {
				connsByKey[key] = map[string]int{}
			}
			connsByKey[key][connID(c)]++
		}
		res.Models[m.ID] = Assignment{Key: key, Name: name, Evidence: ev}
		p := provs[key]
		if p == nil {
			p = &Provider{Key: key, Name: name}
			provs[key] = p
		}
		p.Models++
		if !contains(p.Raw, raw) {
			p.Raw = append(p.Raw, raw)
		}
	}
	for key, p := range provs {
		for c := range connsByKey[key] {
			p.Conns = append(p.Conns, c)
		}
		sort.Strings(p.Conns)
		sort.Strings(p.Raw)
		res.Providers = append(res.Providers, *p)
	}
	sort.Slice(res.Providers, func(i, j int) bool {
		if res.Providers[i].Models != res.Providers[j].Models {
			return res.Providers[i].Models > res.Providers[j].Models
		}
		return res.Providers[i].Key < res.Providers[j].Key
	})

	// Suggestions: distinct keys that share the same connection identifier.
	byConn := map[string]map[string]bool{}
	for key, cs := range connsByKey {
		for c := range cs {
			if byConn[c] == nil {
				byConn[c] = map[string]bool{}
			}
			byConn[c][key] = true
		}
	}
	for c, keys := range byConn {
		if len(keys) < 2 {
			continue
		}
		var ks []string
		for k := range keys {
			ks = append(ks, k)
		}
		sort.Strings(ks)
		res.Suggestions = append(res.Suggestions, Suggestion{Keys: ks, Evidence: "same connection " + c})
	}
	sort.Slice(res.Suggestions, func(i, j int) bool { return res.Suggestions[i].Evidence < res.Suggestions[j].Evidence })
	return res
}

func endpointLabel(name string) string {
	if name == "" {
		return "endpoint"
	}
	return name
}

func coverage(vals map[string]int, n int) float64 {
	if n == 0 {
		return 0
	}
	total := 0
	for _, c := range vals {
		total += c
	}
	return float64(total) / float64(n)
}

// consistency measures how well a candidate field agrees with observed
// connection identifiers: for every field value that has connection
// evidence, the share of models whose connection equals the majority one.
func consistency(models []Model, conns map[string]string, field func(Model) string) (score float64, covered int) {
	if len(conns) == 0 {
		return 1, 0
	}
	per := map[string]map[string]int{}
	for _, m := range models {
		c := connID(conns[m.ID])
		if c == "" {
			continue
		}
		v := field(m)
		if v == "" {
			continue
		}
		if per[v] == nil {
			per[v] = map[string]int{}
		}
		per[v][c]++
		covered++
	}
	if covered == 0 {
		return 1, 0
	}
	agree := 0
	for _, cs := range per {
		best := 0
		for _, n := range cs {
			if n > best {
				best = n
			}
		}
		agree += best
	}
	return float64(agree) / float64(covered), covered
}

// connID reduces "connection/model" to "connection".
func connID(c string) string {
	if i := strings.Index(c, "/"); i > 0 {
		return c[:i]
	}
	return c
}

func rationale(field string, distinct, covered int, cons float64) string {
	s := field + " discriminates between models (" + itoa(distinct) + " distinct values)"
	if covered > 0 {
		s += "; consistent with observed connections for " + itoa(int(cons*100)) + "% of " + itoa(covered) + " models"
	}
	return s
}

func itoa(n int) string { return strconv.Itoa(n) }

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
