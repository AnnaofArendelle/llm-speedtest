// Package modelid normalises model identifiers returned by OpenAI-compatible
// endpoints. Gateways commonly prefix the upstream model name with a provider
// alias ("alias/model") or an organisation ("org/model"); callers frequently
// need the bare name as well as the prefix segments.
package modelid

import (
	"regexp"
	"strings"
)

var bracketRE = regexp.MustCompile(`\[[^\]]*\]`)

// Parts describes the components of a model identifier.
type Parts struct {
	// Raw is the identifier exactly as returned by the endpoint.
	Raw string
	// Name is the last path segment with bracket annotations removed and
	// leading "~" / "@" markers stripped, e.g. "deepseek-v4-flash:free".
	Name string
	// Org joins the path segments between the provider prefix and the name,
	// lower-cased, e.g. "deepseek-ai" for "kilo/deepseek-ai/model".
	Org string
	// Prefix is the first path segment of the raw identifier (before any
	// stripping), or "" when the identifier has no "/".
	Prefix string
	// OwnerStripped reports whether the owned_by prefix was removed.
	OwnerStripped bool
}

// Clean splits a model identifier. When ownedBy is non-empty and the
// identifier starts with "ownedBy/" (case-insensitive), that prefix is
// removed first so that "alias/org/model" yields Org "org" and Name "model".
func Clean(id, ownedBy string) Parts {
	p := Parts{Raw: id}
	s := id
	if i := strings.Index(s, "/"); i > 0 {
		p.Prefix = s[:i]
	}
	if ownedBy != "" && len(s) > len(ownedBy)+1 && strings.EqualFold(s[:len(ownedBy)], ownedBy) && s[len(ownedBy)] == '/' {
		s = s[len(ownedBy)+1:]
		p.OwnerStripped = true
	}
	s = strings.TrimSpace(bracketRE.ReplaceAllString(s, ""))
	var segs []string
	for _, seg := range strings.Split(s, "/") {
		seg = strings.TrimLeft(seg, "~@")
		if seg != "" {
			segs = append(segs, seg)
		}
	}
	if len(segs) == 0 {
		p.Name = s
		return p
	}
	p.Name = segs[len(segs)-1]
	p.Org = strings.ToLower(strings.Join(segs[:len(segs)-1], "-"))
	return p
}

// Name is a convenience wrapper returning only the cleaned name.
func Name(id, ownedBy string) string { return Clean(id, ownedBy).Name }

// Upstream returns the identifier with the owned_by prefix removed but all
// other path segments intact, which is what most gateways forward upstream.
func Upstream(id, ownedBy string) string {
	if ownedBy != "" && len(id) > len(ownedBy)+1 && strings.EqualFold(id[:len(ownedBy)], ownedBy) && id[len(ownedBy)] == '/' {
		return id[len(ownedBy)+1:]
	}
	return id
}
