// Package family groups model identifiers into model families ("series")
// using endpoint-agnostic tokenisation rules, then refines the grouping with
// evidence learned from the endpoint itself (served model names, vendor
// vocabulary) and user overrides. No vendor-specific alias table is built in.
package family

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/AnnaofArendelle/llm-speedtest/internal/modelid"
)

// Family is the normalised identity of a model.
type Family struct {
	Key   string   `json:"key"`   // base|mod1+mod2|ver, the grouping key
	Base  string   `json:"base"`  // product/base name, e.g. "deepseek"
	Mods  []string `json:"mods"`  // tier/speciality words kept in the key, e.g. ["flash"]
	Ver   int      `json:"ver"`   // major version, 0 when none
	HasV  bool     `json:"has_v"` // whether a version was found
	Canon string   `json:"canon"` // canonical display form, e.g. "deepseek-v4-flash"
	Name  string   `json:"name"`  // cleaned model name the key was derived from
	Org   string   `json:"org"`   // organisation segment (lower-case), if any
	// Generic marks routing pseudo-models ("auto", "free", "default", ...)
	// whose backend varies per request; learners never merge them.
	Generic bool `json:"generic,omitempty"`
}

// Rules holds the vocabularies used by the tokeniser. All sets are editable
// so that deployments can adapt to unusual naming schemes.
type Rules struct {
	Tier        map[string]bool   // kept in the key, distinguish product lines (flash/pro/opus...)
	Spec        map[string]string // kept in the key, modality/speciality words (vl/coder/embedding...) with canonical form
	Qual        map[string]bool   // dropped qualifiers (preview/instruct/free/...)
	Effort      map[string]bool   // dropped effort levels when the name has a thinking marker or is a claude model
	VendorLead  map[string]bool   // leading vendor words dropped when followed by a name
	Generic     map[string]bool   // generic names that get prefixed with the org to stay distinct
	VerPrefix   map[string]bool   // single letters that prefix versions (v4, k2, m3, r1, q2)
	SizeLetters map[string]bool   // bare size letters dropped (xs/s/m/l)
}

var (
	sizeRE    = regexp.MustCompile(`^(?:\d+x)?\d+(?:\.\d+)?[bmk]$|^[aex]\d+(?:\.\d+)?b$`)
	verRE     = regexp.MustCompile(`^([a-z]*)(\d+(?:\.\d+)*)([a-z]*)$`)
	splitRE   = regexp.MustCompile(`[-_\s]+`)
	suffixRE  = regexp.MustCompile(`[-_\s:]+`)
	dotNumRE  = regexp.MustCompile(`([a-z])\.(\d)`)
	nonAlnum  = regexp.MustCompile(`[^a-z0-9]+`)
	digitsRE  = regexp.MustCompile(`^\d+$`)
	defaultRu = buildDefault()
)

func set(words ...string) map[string]bool {
	m := make(map[string]bool, len(words))
	for _, w := range words {
		m[w] = true
	}
	return m
}

func buildDefault() *Rules {
	return &Rules{
		Tier: set("flash", "pro", "max", "plus", "lite", "mini", "nano", "micro", "tiny", "small", "medium", "large", "xl", "xxl",
			"ultra", "super", "turbo", "air", "instant", "opus", "sonnet", "haiku", "fable", "flashx", "lightning"),
		Spec: map[string]string{"vl": "vl", "vision": "vision", "omni": "omni", "code": "code", "coder": "coder", "codex": "codex",
			"embedding": "embedding", "embeddings": "embedding", "embed": "embedding", "rerank": "rerank", "reranker": "rerank",
			"tts": "tts", "audio": "audio", "image": "image", "video": "video", "search": "search", "guard": "guard",
			"safeguard": "guard", "translate": "translate", "reasoner": "reasoner", "math": "math", "voice": "voice", "ocr": "ocr",
			"asr": "asr", "research": "research"},
		Qual: set("free", "preview", "exp", "experimental", "beta", "alpha", "rc", "nightly", "latest", "stable", "instruct", "it",
			"chat", "base", "fast", "highspeed", "ultraspeed", "hyperspeed", "speed", "thinking", "think", "reasoning",
			"nonthinking", "non", "uncensored", "abliterated", "antislop", "heretic", "rp", "sdft", "novelist", "cybersecurity",
			"contributor", "open", "distill", "distilled", "terminus", "multilingual", "api", "hd", "exclusive", "nitro", "flat",
			"cloud", "online", "extended", "long", "context", "auto", "new", "old", "dev", "test", "demo", "legacy",
			"derestricted", "venice", "edition", "custom", "ft", "finetune", "merged", "merge", "gguf", "awq", "gptq", "int4",
			"int8", "fp8", "fp16", "bf16", "nvfp4", "fp4", "q4", "q8", "quantized", "moe", "dense", "sft", "dpo", "rlhf", "hf",
			"c4ai", "next", "general", "standard", "default", "official", "en", "zh", "cn", "jp", "ko", "fr", "de", "rl"),
		Effort:      set("low", "medium", "high", "max", "minimal", "xhigh"),
		VendorLead:  set("nvidia", "meta", "google", "openai", "anthropic", "xiaomi", "tencent", "ibm", "amazon", "microsoft", "bytedance", "moonshotai", "alibaba", "zai", "cohere", "mistralai", "deepseekai", "xai"),
		Generic:     set("free", "auto", "default", "balanced", "efficient", "frontier", "small", "large", "medium", "fast", "best"),
		VerPrefix:   set("v", "k", "m", "r", "q"),
		SizeLetters: set("xs", "s", "m", "l"),
	}
}

// DefaultRules returns a copy of the built-in rules.
func DefaultRules() *Rules {
	r := *defaultRu
	return &r
}

// Key derives the family of a model id with the default rules.
func Key(id, ownedBy string) Family { return defaultRu.Key(id, ownedBy) }

// Key derives the family of a model id.
func (r *Rules) Key(id, ownedBy string) Family {
	parts := modelid.Clean(id, ownedBy)
	name := parts.Name
	low := strings.ToLower(name)
	var suffixTokens []string
	if i := strings.Index(low, ":"); i >= 0 {
		suf := low[i+1:]
		low = low[:i]
		for _, st := range suffixRE.Split(suf, -1) {
			if st != "" && sizeRE.MatchString(st) {
				suffixTokens = append(suffixTokens, st)
			}
		}
	}
	low = dotNumRE.ReplaceAllString(low, "$1-$2") // flux.2 -> flux-2
	var tokens []string
	for _, t := range splitRE.Split(low, -1) {
		if t != "" {
			tokens = append(tokens, t)
		}
	}
	tokens = append(tokens, suffixTokens...)
	hasThinking := false
	for _, t := range tokens {
		if t == "thinking" || t == "think" {
			hasThinking = true
		}
	}
	isClaude := len(tokens) > 0 && tokens[0] == "claude"
	var base, mods, canon []string
	ver, hasV := 0, false
	addWord := func(t string) {
		switch {
		case r.Tier[t]:
			mods = append(mods, t)
			canon = append(canon, t)
		case r.Spec[t] != "":
			mods = append(mods, r.Spec[t])
			canon = append(canon, r.Spec[t])
		case !hasV && (len(mods) == 0 || len(base) == 0):
			base = append(base, t)
			canon = append(canon, t)
		}
	}
	for idx, t := range tokens {
		if r.Qual[t] {
			continue
		}
		if (hasThinking || isClaude) && r.Effort[t] {
			continue
		}
		if sizeRE.MatchString(t) {
			continue
		}
		if digitsRE.MatchString(t) {
			if len(t) >= 4 || (len(t) == 2 && t[0] == '0') {
				continue
			}
			if len(t) <= 2 && idx+1 < len(tokens) && digitsRE.MatchString(tokens[idx+1]) && len(tokens[idx+1]) == 4 {
				continue // 12-2024 style dates
			}
			if !hasV {
				n, _ := strconv.Atoi(t)
				ver, hasV = n, true
				canon = append(canon, strconv.Itoa(n))
			}
			continue
		}
		if len(base) == 0 && len(mods) == 0 && r.VendorLead[t] && idx+1 < len(tokens) && (tokens[idx+1][0] < '0' || tokens[idx+1][0] > '9') {
			continue
		}
		if t == "v" {
			mods = append(mods, "vision")
			canon = append(canon, "vision")
			continue
		}
		if r.SizeLetters[t] && len(base) > 0 {
			continue
		}
		if vm := verRE.FindStringSubmatch(t); vm != nil {
			pre, num, suf := vm[1], vm[2], vm[3]
			if (suf == "b" || suf == "m") && pre != "" {
				// letter-prefixed size such as r7b (Command R 7B): keep the letter as part of the base
				if !hasV && len(mods) == 0 && len(pre) <= 2 {
					base = append(base, pre)
					canon = append(canon, pre)
				}
				continue
			}
			majorStr := num
			if i := strings.Index(num, "."); i >= 0 {
				majorStr = num[:i]
			}
			if len(majorStr) > 9 {
				addWord(t)
				continue
			}
			major, _ := strconv.Atoi(majorStr)
			word := ""
			fused := false
			if pre != "" {
				if len(pre) == 1 && r.VerPrefix[pre] && idx > 0 {
					word = ""
				} else {
					word, fused = pre, true
				}
			}
			if word != "" {
				addWord(word)
			}
			if !hasV {
				ver, hasV = major, true
				switch {
				case fused && len(canon) > 0 && canon[len(canon)-1] == word:
					canon[len(canon)-1] = word + strconv.Itoa(major)
				case pre != "" && !fused:
					canon = append(canon, pre+strconv.Itoa(major))
				default:
					canon = append(canon, strconv.Itoa(major))
				}
			}
			if suf == "o" && len(canon) > 0 && canon[len(canon)-1] == strconv.Itoa(major) {
				canon[len(canon)-1] = strconv.Itoa(major) + "o"
			}
			switch {
			case suf == "v":
				mods = append(mods, "vision")
				canon = append(canon, "vision")
			case suf != "" && (r.Tier[suf] || r.Spec[suf] != ""):
				addWord(suf) // e.g. dsv41flash -> tier "flash"
			}
			continue
		}
		addWord(t)
	}
	if len(base) == 0 {
		b := strings.Trim(nonAlnum.ReplaceAllString(low, "-"), "-")
		if b == "" {
			b = low
		}
		base = []string{b}
		if len(canon) == 0 {
			canon = []string{b}
		}
	}
	generic := r.Generic[base[0]]
	if generic && parts.Org != "" {
		base = append([]string{parts.Org}, base...)
		canon = append([]string{parts.Org}, canon...)
	}
	seen := map[string]bool{}
	var umods []string
	for _, m := range mods {
		if !seen[m] {
			seen[m] = true
			umods = append(umods, m)
		}
	}
	f := mk(strings.Join(base, "-"), umods, ver, hasV, name, strings.Join(canon, "-"), parts.Org)
	f.Generic = generic
	return f
}

func mk(base string, mods []string, ver int, hasV bool, name, canon, org string) Family {
	v := ""
	if hasV {
		v = strconv.Itoa(ver)
	}
	if mods == nil {
		mods = []string{}
	}
	return Family{Key: base + "|" + strings.Join(mods, "+") + "|" + v, Base: base, Mods: mods, Ver: ver, HasV: hasV, Canon: canon, Name: name, Org: org}
}

// KeyOf builds a key from components (used by the learners).
func KeyOf(base string, mods []string, ver int, hasV bool) string {
	return mk(base, mods, ver, hasV, "", "", "").Key
}

// Display picks the display name of a family from its members' canonical
// forms: the most common one, then the shortest, then lexicographic.
func Display(canons []string) string {
	if len(canons) == 0 {
		return ""
	}
	counts := map[string]int{}
	for _, c := range canons {
		counts[c]++
	}
	type kv struct {
		c string
		n int
	}
	var list []kv
	for c, n := range counts {
		list = append(list, kv{c, n})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].n != list[j].n {
			return list[i].n > list[j].n
		}
		if len(list[i].c) != len(list[j].c) {
			return len(list[i].c) < len(list[j].c)
		}
		return list[i].c < list[j].c
	})
	return list[0].c
}
