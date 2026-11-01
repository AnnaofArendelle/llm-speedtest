package family

import (
	"sort"
	"strconv"
	"strings"
)

// Model is the minimal input the learners need.
type Model struct {
	ID            string
	OwnedBy       string
	ContextLength int
}

// Source explains how an assignment was decided.
const (
	SourceRule   = "rule"     // tokenisation rules only
	SourceServed = "served"   // the upstream answered with a differently named model
	SourceAbbrev = "abbrev"   // a short id was expanded against the endpoint's own vocabulary
	SourceUser   = "override" // a user override
)

// Assignment is the final family of one model together with its evidence.
type Assignment struct {
	Key      string   `json:"key"`
	Canon    string   `json:"canon"`
	Source   string   `json:"source"`
	Evidence string   `json:"evidence,omitempty"`
	Rule     Family   `json:"-"`
	Members  []string `json:"-"`
}

// Override pins a model to a family key chosen by the user.
type Override struct {
	Key   string
	Canon string
}

// Learn assigns every model to a family. served maps a model id to the model
// name the upstream reported in a successful response (if any); overrides
// take precedence over everything else.
func Learn(r *Rules, models []Model, served map[string]string, overrides map[string]Override) map[string]Assignment {
	if r == nil {
		r = defaultRu
	}
	out := make(map[string]Assignment, len(models))
	rule := make(map[string]Family, len(models))
	members := map[string][]string{}
	for _, m := range models {
		f := r.Key(m.ID, m.OwnedBy)
		rule[m.ID] = f
		members[f.Key] = append(members[f.Key], m.ID)
		out[m.ID] = Assignment{Key: f.Key, Canon: f.Canon, Source: SourceRule, Rule: f}
	}
	canonOf := func(key string) string {
		var cs []string
		for _, id := range members[key] {
			cs = append(cs, rule[id].Canon)
		}
		return Display(cs)
	}

	// 1. Served model names: an alias without a version that the upstream
	// resolves to a versioned model joins that model's family.
	for _, m := range models {
		s, ok := served[m.ID]
		if !ok || s == "" {
			continue
		}
		req := rule[m.ID]
		if req.HasV || req.Generic {
			continue
		}
		srv := r.Key(s, "")
		if !srv.HasV || srv.Key == req.Key {
			continue
		}
		canon := srv.Canon
		if len(members[srv.Key]) > 0 {
			canon = canonOf(srv.Key)
		}
		out[m.ID] = Assignment{Key: srv.Key, Canon: canon, Source: SourceServed, Evidence: "served as " + s, Rule: req}
	}

	// 2. Abbreviations: build the vocabulary of well-populated bases and
	// expand short, otherwise unknown bases against it (unique match only).
	vocab := map[string]int{}
	for key, ids := range members {
		base := strings.SplitN(key, "|", 2)[0]
		if len(base) >= 4 {
			vocab[base] += len(ids)
		}
	}
	for _, m := range models {
		a := out[m.ID]
		if a.Source != SourceRule {
			continue
		}
		req := rule[m.ID]
		base := req.Base
		if req.Generic || len(base) > 3 || vocab[base] >= 2 || strings.Contains(base, "-") {
			continue
		}
		cands := abbrevCandidates(base, req.HasV, vocab)
		if len(cands) == 0 {
			continue
		}
		vers := []int{}
		if req.HasV {
			vers = append(vers, req.Ver)
			if req.Ver >= 10 {
				if d, err := strconv.Atoi(strconv.Itoa(req.Ver)[:1]); err == nil {
					vers = append(vers, d)
				}
			}
		}
		var matches []string
		for _, v := range cands {
			for _, ver := range vers {
				k := KeyOf(v, req.Mods, ver, true)
				if len(members[k]) > 0 && !contains(matches, k) {
					matches = append(matches, k)
				}
			}
		}
		if len(matches) == 1 {
			out[m.ID] = Assignment{Key: matches[0], Canon: canonOf(matches[0]), Source: SourceAbbrev, Evidence: "abbreviation of " + canonOf(matches[0]), Rule: req}
		}
	}

	// 3. User overrides.
	for id, ov := range overrides {
		if a, ok := out[id]; ok {
			canon := ov.Canon
			if canon == "" {
				if len(members[ov.Key]) > 0 {
					canon = canonOf(ov.Key)
				} else {
					canon = strings.ReplaceAll(strings.ReplaceAll(ov.Key, "|", "-"), "+", "-")
					canon = strings.Trim(strings.ReplaceAll(canon, "--", "-"), "-")
				}
			}
			out[id] = Assignment{Key: ov.Key, Canon: canon, Source: SourceUser, Evidence: "user override", Rule: a.Rule}
		}
	}
	return out
}

// abbrevCandidates returns vocabulary bases that the short base abbreviates:
// same first letter and the short form is a subsequence of the long form. A
// trailing "v" (as in "dsv4") is ignored when a version follows.
func abbrevCandidates(short string, hasV bool, vocab map[string]int) []string {
	forms := []string{short}
	if hasV && len(short) > 1 && strings.HasSuffix(short, "v") {
		forms = append(forms, strings.TrimSuffix(short, "v"))
	}
	var out []string
	for long := range vocab {
		for _, f := range forms {
			if len(f) >= 1 && f[0] == long[0] && isSubsequence(f, long) {
				out = append(out, long)
				break
			}
		}
	}
	sort.Strings(out)
	return out
}

func isSubsequence(short, long string) bool {
	if short == "" {
		return false
	}
	i := 0
	for j := 0; j < len(long) && i < len(short); j++ {
		if long[j] == short[i] {
			i++
		}
	}
	return i == len(short)
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
