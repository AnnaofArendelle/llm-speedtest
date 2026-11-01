// Package report renders analysis results as Markdown, JSON and CSV. The
// Markdown layout mirrors the dashboard: summary, methodology, providers,
// rankings, failure inventory, family details and caveats.
package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/analyze"
	"github.com/AnnaofArendelle/llm-speedtest/internal/classify"
)

// Meta describes the run being reported.
type Meta struct {
	Title       string
	Endpoint    string
	Dialect     string
	Version     string // tool version
	Generated   time.Time
	Started     time.Time
	Finished    time.Time
	Preset      string
	RoundsDesc  string // human description of the round plan
	Rounds      []RoundLine
	ExcludeNote string // e.g. "excluded 14 combo models"
	Lang        string // zh (default) | en
	Extra       []Section
}

// RoundLine summarises one round for the methodology table.
type RoundLine struct {
	Round   int
	Kind    string
	Targets int
	OK      int
	Fail    int
}

// Section is an additional pre-rendered Markdown block (e.g. latency lab).
type Section struct {
	Title string
	Body  string
}

type strs struct {
	lang string
}

func (s strs) t(zh, en string) string {
	if s.lang == "en" {
		return en
	}
	return zh
}

func esc(v any) string {
	str := fmt.Sprint(v)
	str = strings.ReplaceAll(str, "\\", "\\\\")
	str = strings.ReplaceAll(str, "|", "\\|")
	str = strings.ReplaceAll(str, "\n", " ")
	str = strings.ReplaceAll(str, "\r", " ")
	return strings.TrimSpace(str)
}

func trunc(v string, n int) string {
	v = esc(v)
	r := []rune(v)
	if len(r) <= n {
		return v
	}
	return string(r[:n-1]) + "…"
}

func fs(x float64) string {
	if x <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.2fs", x)
}

func pct(a, b int) string {
	if b == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f%%", 100*float64(a)/float64(b))
}

func code(v string) string { return "`" + strings.ReplaceAll(v, "`", "") + "`" }

func fctx(n int) string {
	switch {
	case n <= 0:
		return "-"
	case n >= 1000:
		return fmt.Sprintf("%dk", n/1000)
	}
	return fmt.Sprintf("%d", n)
}

func reasonName(code string, lang string) string {
	info := classify.Info(code)
	if lang == "en" {
		return info.NameEN
	}
	return info.Name
}

func tier(ms *analyze.ModelStat, lang string) string {
	return fmt.Sprintf("%s（%d/%d）", analyze.TierName(ms.Tier, lang), ms.OK, ms.Attempts)
}

func ftier(f *analyze.FamilyStat, lang string) string {
	return fmt.Sprintf("%s（%d/%d）", analyze.TierName(f.Tier, lang), f.Successes, f.Attempts)
}

func sortedKeys(m map[string]int) []string {
	type kv struct {
		k string
		v int
	}
	var list []kv
	for k, v := range m {
		list = append(list, kv{k, v})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].v != list[j].v {
			return list[i].v > list[j].v
		}
		return list[i].k < list[j].k
	})
	out := make([]string, 0, len(list))
	for _, e := range list {
		out = append(out, e.k)
	}
	return out
}

func joinCounts(m map[string]int, lang string, max int, reasons bool) string {
	var parts []string
	for i, k := range sortedKeys(m) {
		if max > 0 && i >= max {
			break
		}
		name := k
		if reasons {
			name = reasonName(k, lang)
		}
		parts = append(parts, fmt.Sprintf("%s(%d)", esc(name), m[k]))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, ", ")
}

// Markdown renders the full scan report.
func Markdown(res *analyze.Result, meta Meta) string {
	s := strs{lang: meta.Lang}
	var b strings.Builder
	w := func(format string, args ...any) {
		if len(args) == 0 {
			b.WriteString(format)
		} else {
			fmt.Fprintf(&b, format, args...)
		}
		b.WriteString("\n")
	}
	sum := res.Summary
	title := meta.Title
	if title == "" {
		title = s.t("LLM API 测速报告", "LLM API Speed Test Report")
	}
	gen := meta.Generated
	if gen.IsZero() {
		gen = time.Now()
	}
	w("# %s", title)
	w("")
	w("- %s%s", s.t("生成时间：", "Generated: "), gen.Format("2006-01-02 15:04:05"))
	if !meta.Started.IsZero() {
		w("- %s%s ～ %s", s.t("测试窗口：", "Test window: "), meta.Started.Format("2006-01-02 15:04:05"), meta.Finished.Format("2006-01-02 15:04:05"))
	}
	w("- %s%s%s", s.t("端点：", "Endpoint: "), code(meta.Endpoint), dialectNote(meta, s))
	scope := fmt.Sprintf(s.t("共 **%d** 个模型，%d 个提供商，归并为 **%d** 个模型系列", "**%d** models, %d providers, grouped into **%d** families"), sum.Models, sum.Providers, sum.Families)
	if meta.ExcludeNote != "" {
		scope += "（" + meta.ExcludeNote + "）"
	}
	w("- %s%s", s.t("范围：", "Scope: "), scope)
	if len(meta.Rounds) > 0 {
		var rs []string
		for _, r := range meta.Rounds {
			rs = append(rs, fmt.Sprintf(s.t("第 %d 轮 %d 次（成功 %d）", "round %d: %d requests (%d ok)"), r.Round, r.OK+r.Fail, r.OK))
		}
		w("- %s%s", s.t("测试轮次：", "Rounds: "), strings.Join(rs, s.t("；", "; ")))
	}
	if meta.Version != "" {
		w("- %sllm-speedtest %s", s.t("工具版本：", "Tool: "), meta.Version)
	}
	w("")

	// ---------- 1. Summary ----------
	w("## 1. %s", s.t("结论摘要", "Summary"))
	w("")
	w("| %s | %s |", s.t("指标", "Metric"), s.t("数值", "Value"))
	w("|---|---|")
	w("| %s | %d |", s.t("测试模型数", "Models tested"), sum.Models)
	w("| %s | **%d**（%s） |", s.t("可用（至少一次成功返回）", "Available (at least one success)"), sum.OK, pct(sum.OK, sum.Models))
	w("| %s | %d（%s） |", s.t("不可用", "Unavailable"), sum.Fail, pct(sum.Fail, sum.Models))
	w("| %s | %d（%s） |", s.t("其中因提供商账号问题不可用（无凭证/认证失败/余额或额度用尽/套餐限制）", "…of which provider/account-level (credentials, auth, balance, quota, plan)"), sum.ProviderLevelFail, pct(sum.ProviderLevelFail, sum.Fail))
	w("| %s | %d（%s %d，%s %d） |", s.t("模型系列数", "Families"), sum.Families, s.t("有可用成员", "with an available member"), sum.FamiliesOK, s.t("全员不可用", "entirely unavailable"), sum.FamiliesUnavail)
	w("| %s | A %d，B %d，C %d |", s.t("可用模型的稳定性（A 多轮全部成功，B 成功率 ≥ 2/3，C 不稳定）", "Stability of available models (A all rounds ok, B ≥ 2/3, C unstable)"), sum.TierCounts["A"], sum.TierCounts["B"], sum.TierCounts["C"])
	w("| %s | A %d，B %d，C %d |", s.t("有可用成员的系列的稳定性", "Stability of available families"), sum.FamilyTierCounts["A"], sum.FamilyTierCounts["B"], sum.FamilyTierCounts["C"])
	if sum.OK > 0 {
		w("| %s | %s / %s / %s / %s |", s.t("可用模型延迟中位数 / P90 / 最快 / 最慢", "Latency median / P90 / min / max"), fs(sum.LatencyMedian), fs(sum.LatencyP90), fs(sum.LatencyMin), fs(sum.LatencyMax))
	}
	w("")
	w("%s", s.t("延迟口径：非流式 `chat/completions` 从发出请求到收到完整响应的耗时（`max_tokens` 很小），约等于首 token 时间加路由开销，反映的是客户端实际可感知的响应速度，不是提供商的纯推理延迟。",
		"Latency is the end-to-end time of a non-streaming `chat/completions` request with a tiny `max_tokens`, i.e. roughly time-to-first-token plus routing overhead as perceived by the client, not the provider's pure inference latency."))
	w("")
	if len(res.Available) > 0 {
		w("### 1.1 %s", s.t("最稳定且最快的 20 个可用模型", "Top 20 most stable and fastest models"))
		w("")
		w("| # | %s | %s | %s | %s | %s | %s | %s |", s.t("模型 ID", "Model"), s.t("提供商", "Provider"), s.t("系列", "Family"), s.t("稳定性（成功/尝试）", "Stability (ok/attempts)"), s.t("中位", "Median"), s.t("最小", "Min"), s.t("最大", "Max"))
		w("|---|---|---|---|---|---|---|---|")
		for i, m := range res.Available {
			if i >= 20 {
				break
			}
			w("| %d | %s | %s | %s | %s | %s | %s | %s |", i+1, code(m.ID), esc(m.ProviderName), esc(m.FamilyCanon), tier(m, s.lang), fs(m.Med), fs(m.Min), fs(m.Max))
		}
		w("")
		w("### 1.2 %s", s.t("可用模型延迟分布", "Latency distribution of available models"))
		w("")
		w("| %s | %s | %s |", s.t("中位延迟区间", "Median latency"), s.t("模型数", "Models"), s.t("占可用", "Share"))
		w("|---|---|---|")
		for _, bk := range sum.Buckets {
			w("| %s | %d | %s |", bk.Label, bk.Count, pct(bk.Count, sum.OK))
		}
		w("")
	}
	w("### 1.3 %s", s.t("不可用原因分布", "Failure reasons"))
	w("")
	w("| %s | %s | %s | %s |", s.t("原因", "Reason"), s.t("模型数", "Models"), s.t("占不可用", "Share"), s.t("层级", "Level"))
	w("|---|---|---|---|")
	for _, c := range classify.Order {
		if n := sum.Reasons[c]; n > 0 {
			lvl := s.t("模型级/网络级", "model / network")
			if classify.IsProviderLevel(c) {
				lvl = s.t("提供商/账号级", "provider / account")
			}
			w("| %s | %d | %s | %s |", reasonName(c, s.lang), n, pct(n, sum.Fail), lvl)
		}
	}
	w("")
	if len(res.Families) > 0 {
		w("### 1.4 %s", s.t("综合排名前 15 的模型系列（可用提供商多、稳定、快 → 少、不稳定、慢）", "Top 15 families by composite score"))
		w("")
		w(s.t("综合分 = 稳定性分 × %.1f + 提供商广度分 × %.1f + 速度分 × %.1f，详见第 4 节。", "Composite = stability × %.1f + provider breadth × %.1f + speed × %.1f; see section 4."), res.Scoring.WStab, res.Scoring.WBreadth, res.Scoring.WSpeed)
		w("")
		w("| # | %s | %s | %s | %s | %s | %s | %s |", s.t("系列", "Family"), s.t("综合分", "Score"), s.t("可用提供商数", "Providers"), s.t("可用成员/总成员", "Available/total"), s.t("稳定性（成功/尝试）", "Stability"), s.t("系列中位", "Median"), s.t("最稳定且最快的成员 @ 提供商", "Best member @ provider"))
		w("|---|---|---|---|---|---|---|---|")
		for i, f := range res.Families {
			if i >= 15 {
				break
			}
			w("| %d | **%s** | %.1f | %d（%s %d） | %d/%d | %s | %s | %s @ %s：%s |", i+1, esc(f.Display), f.Score, f.NProvOK, s.t("稳定", "stable"), f.NProvStable, f.OK, f.N, ftier(f, s.lang), fs(f.FMed), code(f.Best.ID), esc(f.Best.ProviderName), fs(f.Best.Med))
		}
		w("")
	}

	// ---------- 2. Method ----------
	w("## 2. %s", s.t("测试方法", "Method"))
	w("")
	w("1. %s", s.t("调用 `GET /models` 取得完整模型列表（含 `owned_by` 与上下文元数据）。", "Fetch the complete model list with `GET /models` (including `owned_by` and context metadata)."))
	w("2. %s", s.t("**第 1 轮**：对每个模型发送一次最小化 `chat/completions` 请求，记录 HTTP 状态、嵌套的上游状态码、报错原文、延迟与实际返回的模型名；全局与每提供商并发均受限。", "**Round 1**: one minimal `chat/completions` request per model, recording HTTP status, nested upstream status, error text, latency and the served model name; concurrency is bounded globally and per provider."))
	w("3. %s", s.t("**复测轮**：对瞬时性失败（超时、限流、上游 5xx、无可用渠道/容量、空响应、网络错误）以更低并发、更长超时再测一次；若网关给出 `reset after`/`retry-after` 提示则等待到期后再试。认证失败、无凭证、余额不足、模型不存在等确定性错误不复测。", "**Retry round**: transient failures (timeouts, rate limits, upstream 5xx, no channel/capacity, empty responses, network errors) are retried with lower concurrency and a longer timeout, honouring `reset after`/`retry-after` hints. Deterministic failures (auth, credentials, balance, missing model) are not retried."))
	w("4. %s", s.t("**补采样轮**：对所有成功过的模型再各测若干次，排序使用延迟**中位数**，并给出最小/最大值。", "**Resample rounds**: every model that succeeded is probed again; rankings use the **median** latency with min/max shown."))
	w("5. %s", s.t("**最终确认轮**：对复测后仍属瞬时性失败的模型再测一次；仍失败才判定为不可用。", "**Final round**: remaining transient failures get one last attempt before being declared unavailable."))
	w("6. %s", s.t("成功判定：HTTP 200 且响应体包含 `choices`。失败原因按状态码优先、关键词其次归类，报错原文全部保留。", "Success means HTTP 200 with a `choices` array. Failures are classified status-code first, keywords second; the original message is always kept."))
	w("7. %s", s.t("**提供商归属**：不使用任何内置名称表，而是根据本端点返回数据动态选择最能区分模型且与错误中连接标识一致的字段（`owned_by` / id 前缀），可在界面中重命名或合并。", "**Provider attribution** uses no built-in table: the field that best discriminates models and agrees with connection identifiers seen in errors (`owned_by` / id prefix) is chosen per endpoint; names can be edited in the UI."))
	w("8. %s", s.t("**系列归并**：去掉提供商/组织前缀与 `:free` 类后缀、参数规模、日期戳、量化标记与修饰词，版本只保留主版本号，保留区分产品线的档位/特长词；别名通过上游实际返回的模型名与本端点自身的词表学习，不使用内置别名字典。", "**Family grouping** strips provider/org prefixes, `:free`-style suffixes, parameter sizes, date stamps, quantisation markers and qualifiers, keeps only the major version and the tier/speciality words that distinguish product lines; aliases are learned from served model names and the endpoint's own vocabulary, never from a built-in dictionary."))
	w("9. %s", s.t("**排名**：先按稳定性等级（A = 多轮尝试全部成功，B = 成功率 ≥ 2/3，C = 成功率 < 2/3），同级内按中位延迟；系列另给出综合分 = 稳定性 × 权重 + 提供商广度 × 权重 + 速度（对数刻度）× 权重。", "**Ranking**: stability tier first (A = every attempt succeeded, B = success rate ≥ 2/3, C = below), then median latency; families additionally get a composite score = stability × w + provider breadth × w + log-scaled speed × w."))
	w("")
	if res.ProviderRationale != "" {
		w("%s%s（%s）", s.t("本端点的提供商字段：", "Provider field for this endpoint: "), code(res.ProviderField), esc(res.ProviderRationale))
		w("")
	}

	// ---------- 3. Providers ----------
	w("## 3. %s", s.t("提供商总览", "Providers"))
	w("")
	w("| %s | %s | %s | %s | %s | %s | %s | %s |", s.t("提供商", "Provider"), s.t("模型数", "Models"), s.t("系列数", "Families"), s.t("可用", "Available"), s.t("可用率", "Rate"), s.t("可用模型中位延迟", "Median latency"), s.t("主要失败原因", "Top failure reason"), s.t("诊断", "Diagnosis"))
	w("|---|---|---|---|---|---|---|---|")
	for _, p := range res.Providers {
		diag := s.t("全部可用", "all available")
		top := "-"
		if p.TopReason != "" {
			top = reasonName(p.TopReason, s.lang)
			diag = fmt.Sprintf(s.t("%s：%d/%d 个失败模型", "%s: %d/%d failed models"), top, p.TopReasonCount, p.Models-p.OK)
			if p.AccountLevelAll {
				diag += s.t("；**整个提供商账号层面不可用**", "; **entire provider unavailable at account level**")
			}
		}
		w("| %s | %d | %d | %d | %s | %s | %s | %s |", esc(p.Name), p.Models, p.Families, p.OK, pct(p.OK, p.Models), fs(p.Med), top, diag)
	}
	w("")

	// ---------- 4. Rankings ----------
	w("## 4. %s", s.t("模型系列测速总排名", "Family rankings"))
	w("")
	w("%s", s.t("**4.1 综合排名**：从“可用提供商最多、响应最快、最稳定”排到“最少、最慢、最不稳定”。", "**4.1 Composite ranking**: from most providers / fastest / most stable to fewest / slowest / least stable."))
	w("")
	w("| %s | %s | %s |", s.t("分项", "Component"), s.t("计算方式", "Definition"), s.t("权重", "Weight"))
	w("|---|---|---|")
	w("| %s | %s | %.2f |", s.t("稳定性分", "Stability"), s.t("该系列可用成员全部尝试的成功率 × 100", "success rate over all attempts of the family's available members × 100"), res.Scoring.WStab)
	w("| %s | %s | %.2f |", s.t("提供商广度分", "Provider breadth"), s.t("有可用成员的提供商数 ÷ 全部系列中的最大值 × 100", "providers with an available member ÷ the maximum across families × 100"), res.Scoring.WBreadth)
	w("| %s | %s | %.2f |", s.t("速度分", "Speed"), fmt.Sprintf(s.t("系列中位延迟按对数刻度映射：≤ %.1f s 得 100 分，≥ %.0f s 得 0 分", "median latency on a log scale: ≤ %.1f s scores 100, ≥ %.0f s scores 0"), res.Scoring.SpeedLo, res.Scoring.SpeedHi), res.Scoring.WSpeed)
	w("")
	w("%s", s.t("稳定性只统计“有过成功”的成员的全部尝试；提供商账号层面的失败不反映模型本身，不计入系列稳定性，但在“可用成员/总成员”和“不可用成员主因”中体现。", "Stability counts every attempt of members that succeeded at least once; provider/account-level failures say nothing about the model and are excluded from family stability but shown in the member counts and failure reasons."))
	w("")
	w("### 4.1 %s", s.t("综合排名", "Composite ranking"))
	w("")
	w("| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |", s.t("排名", "#"), s.t("系列", "Family"), s.t("综合分", "Score"), s.t("稳定性分", "Stab."), s.t("广度分", "Breadth"), s.t("速度分", "Speed"), s.t("可用提供商数（稳定）", "Providers (stable)"), s.t("可用成员/总成员", "Avail./total"), s.t("稳定性（成功/尝试）", "Stability"), s.t("系列中位", "Median"), s.t("最快", "Min"), s.t("最慢", "Max"), s.t("稳定可用的成员（模型 @ 提供商：中位延迟，成功/尝试）", "Stable members (model @ provider: median, ok/attempts)"))
	w("|---|---|---|---|---|---|---|---|---|---|---|---|---|")
	for _, f := range res.Families {
		top := f.Stable
		if len(top) == 0 {
			for _, m := range f.Members {
				if m.Status == "ok" {
					top = append(top, m)
				}
			}
		}
		var parts []string
		for i, m := range top {
			if i >= 3 {
				parts = append(parts, fmt.Sprintf(s.t("…另有 %d 个可用成员见第 6 节", "…%d more in section 6"), f.OK-3))
				break
			}
			parts = append(parts, fmt.Sprintf("%s @ %s：%s（%d/%d）", code(m.ID), esc(m.ProviderName), fs(m.Med), m.OK, m.Attempts))
		}
		w("| %d | **%s** | **%.1f** | %.0f | %.0f | %.0f | %d（%d） | %d/%d | %s | %s | %s | %s | %s |", f.RankComposite, esc(f.Display), f.Score, f.SStab, f.SBreadth, f.SSpeed, f.NProvOK, f.NProvStable, f.OK, f.N, ftier(f, s.lang), fs(f.FMed), fs(f.FMin), fs(f.FMax), strings.Join(parts, "<br>"))
	}
	w("")
	w("### 4.2 %s", s.t("按稳定性等级与延迟排序的系列", "Families by stability tier and latency"))
	w("")
	w("| %s | %s | %s | %s | %s | %s | %s | %s |", s.t("排名", "#"), s.t("系列", "Family"), s.t("稳定性（成功/尝试）", "Stability"), s.t("成员", "Members"), s.t("可用成员", "Available"), s.t("系列中位", "Median"), s.t("可用的提供商", "Available providers"), s.t("不可用成员主因", "Top failure reasons"))
	w("|---|---|---|---|---|---|---|---|")
	for _, f := range res.FamiliesByStability {
		w("| %d | **%s** | %s | %d | %d | %s | %s | %s |", f.RankStability, esc(f.Display), ftier(f, s.lang), f.N, f.OK, fs(f.FMed), joinCounts(f.OKProviders, s.lang, 0, false), joinCounts(f.Reasons, s.lang, 3, true))
	}
	w("")
	w("### 4.3 %s", s.t("全员不可用的系列", "Entirely unavailable families"))
	w("")
	w("| %s | %s | %s | %s |", s.t("系列", "Family"), s.t("成员", "Members"), s.t("主要原因", "Top reasons"), s.t("涉及提供商", "Providers"))
	w("|---|---|---|---|")
	for _, f := range res.UnavailableFamilies {
		w("| **%s** | %d | %s | %s |", esc(f.Display), f.N, joinCounts(f.Reasons, s.lang, 3, true), joinCounts(f.Providers, s.lang, 0, false))
	}
	w("")
	w("### 4.4 %s", s.t("全部可用模型排名（可用性高→低，快→慢）", "All available models (stability, then speed)"))
	w("")
	w("| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |", s.t("排名", "#"), s.t("模型 ID", "Model"), s.t("提供商", "Provider"), s.t("系列", "Family"), s.t("稳定性（成功/尝试）", "Stability"), s.t("中位", "Median"), s.t("最小", "Min"), s.t("最大", "Max"), s.t("波动（最大/最小）", "Jitter (max/min)"), s.t("实际返回模型名", "Served model"), s.t("备注", "Notes"))
	w("|---|---|---|---|---|---|---|---|---|---|---|")
	for i, m := range res.Available {
		w("| %d | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |", i+1, code(m.ID), esc(m.ProviderName), esc(m.FamilyCanon), tier(m, s.lang), fs(m.Med), fs(m.Min), fs(m.Max), jitter(m.Jitter), trunc(m.Served, 40), flakyNote(m, s))
	}
	w("")

	// ---------- 5. Failures ----------
	w("## 5. %s", s.t("不可用模型详细清单", "Unavailable models"))
	w("")
	w("### 5.1 %s", s.t("错误类别说明", "Failure categories"))
	w("")
	w("| %s | %s | %s |", s.t("类别", "Category"), s.t("含义", "Meaning"), s.t("处理建议", "Suggested action"))
	w("|---|---|---|")
	for _, c := range classify.Order {
		if sum.Reasons[c] == 0 {
			continue
		}
		info := classify.Info(c)
		if s.lang == "en" {
			w("| **%s** | %s | %s |", info.NameEN, esc(info.DescriptionEN), esc(info.FixEN))
		} else {
			w("| **%s** | %s | %s |", info.Name, esc(info.Description), esc(info.Fix))
		}
	}
	w("")
	w("%s", s.t("说明：报错原文为网关转发的上游响应（已去掉网关的 `[连接/模型] [状态码]:` 包装）。“HTTP” 列为 `端点返回 / 上游返回`。", "Note: messages are the upstream responses forwarded by the gateway with the `[connection/model] [status]:` wrapper removed. The HTTP column shows `endpoint / upstream`."))
	w("")
	byReason := map[string][]*analyze.ModelStat{}
	for _, m := range res.Models {
		if m.Status != "ok" {
			byReason[m.Reason] = append(byReason[m.Reason], m)
		}
	}
	sec := 2
	for _, c := range classify.Order {
		ms := byReason[c]
		if len(ms) == 0 {
			continue
		}
		w("### 5.%d %s（%d）", sec, reasonName(c, s.lang), len(ms))
		sec++
		w("")
		byProv := map[string][]*analyze.ModelStat{}
		for _, m := range ms {
			byProv[m.ProviderName] = append(byProv[m.ProviderName], m)
		}
		provs := make([]string, 0, len(byProv))
		for p := range byProv {
			provs = append(provs, p)
		}
		sort.Slice(provs, func(i, j int) bool {
			if len(byProv[provs[i]]) != len(byProv[provs[j]]) {
				return len(byProv[provs[i]]) > len(byProv[provs[j]])
			}
			return provs[i] < provs[j]
		})
		for _, p := range provs {
			pms := byProv[p]
			sort.Slice(pms, func(i, j int) bool { return strings.ToLower(pms[i].ID) < strings.ToLower(pms[j].ID) })
			w("#### %s（%d）", esc(p), len(pms))
			w("")
			msgs := map[string]int{}
			for _, m := range pms {
				msgs[trunc(m.Message, 220)]++
			}
			if len(msgs) == 1 && len(pms) > 1 {
				for k := range msgs {
					w(s.t("全部 %d 个模型报错相同：`%s`", "All %d models report the same error: `%s`"), len(pms), k)
				}
				w("")
			}
			w("| %s | HTTP | %s | %s |", s.t("模型 ID", "Model"), s.t("耗时", "Elapsed"), s.t("报错原文", "Message"))
			w("|---|---|---|---|")
			for _, m := range pms {
				w("| %s | %s | %s | %s |", code(m.ID), httpStr(m, s), fs(m.FailLatency), trunc(m.Message, 160))
			}
			w("")
		}
	}

	// ---------- 6. Family details ----------
	w("## 6. %s", s.t("模型系列明细（含每个模型的提供商与测速结果）", "Family details (every model with its provider and results)"))
	w("")
	w("%s", s.t("顺序与第 4.1 节综合排名一致：先有可用成员的系列，再全员不可用的系列。“备注”给出失败原因与报错摘要，或成功时曾失败后复测成功的情况。", "Same order as the composite ranking in 4.1: available families first, then entirely unavailable ones. Notes show the failure reason and message, or that an available model failed in some rounds."))
	w("")
	idx := 1
	writeFamily := func(f *analyze.FamilyStat) {
		w("### %d. %s", idx, esc(f.Display))
		idx++
		w("")
		line := fmt.Sprintf(s.t("归并键 `%s` ｜ 成员 %d，可用 %d ｜ 提供商：%s", "Key `%s` | members %d, available %d | providers: %s"), f.Key, f.N, f.OK, joinCounts(f.Providers, s.lang, 0, false))
		if f.OK > 0 {
			line += fmt.Sprintf(s.t(" ｜ 综合分 %.1f（稳定 %.0f / 广度 %.0f / 速度 %.0f） ｜ 可用提供商 %d 个，其中稳定 %d 个 ｜ 稳定性 %s ｜ 延迟 最快 %s / 中位 %s / 最慢 %s", " | score %.1f (stability %.0f / breadth %.0f / speed %.0f) | %d available providers, %d stable | stability %s | latency min %s / median %s / max %s"), f.Score, f.SStab, f.SBreadth, f.SSpeed, f.NProvOK, f.NProvStable, ftier(f, s.lang), fs(f.FMin), fs(f.FMed), fs(f.FMax))
		}
		if n := f.Sources["served"] + f.Sources["abbrev"] + f.Sources["override"]; n > 0 {
			line += fmt.Sprintf(s.t(" ｜ %d 个成员通过别名学习/覆盖归入", " | %d members joined via alias learning/overrides"), n)
		}
		w("%s", line)
		w("")
		w("| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |", s.t("模型 ID", "Model"), s.t("提供商", "Provider"), s.t("状态", "Status"), s.t("稳定性（成功/尝试）", "Stability"), s.t("中位", "Median"), s.t("最小", "Min"), s.t("最大", "Max"), s.t("实际返回模型名", "Served model"), s.t("上下文", "Context"), s.t("备注", "Notes"))
		w("|---|---|---|---|---|---|---|---|---|---|")
		members := make([]*analyze.ModelStat, len(f.Members))
		copy(members, f.Members)
		sort.Slice(members, func(i, j int) bool {
			a, b := members[i], members[j]
			if (a.Status == "ok") != (b.Status == "ok") {
				return a.Status == "ok"
			}
			if a.Status == "ok" {
				if a.Tier != b.Tier {
					return a.Tier < b.Tier
				}
				if a.Med != b.Med {
					return a.Med < b.Med
				}
			}
			return strings.ToLower(a.ID) < strings.ToLower(b.ID)
		})
		for _, m := range members {
			if m.Status == "ok" {
				note := flakyNote(m, s)
				if m.FamilySource != "" && m.FamilySource != "rule" {
					note = strings.TrimSpace(note + " " + esc(m.FamilyEvidence))
				}
				w("| %s | %s | ✅ %s | %s | %s | %s | %s | %s | %s | %s |", code(m.ID), esc(m.ProviderName), s.t("可用", "ok"), tier(m, s.lang), fs(m.Med), fs(m.Min), fs(m.Max), trunc(m.Served, 40), fctx(m.ContextLength), note)
			} else {
				w("| %s | %s | ❌ %s | 0/%d | - | - | - | - | %s | HTTP %s：%s |", code(m.ID), esc(m.ProviderName), reasonName(m.Reason, s.lang), m.Attempts, fctx(m.ContextLength), httpStr(m, s), trunc(m.Message, 120))
			}
		}
		w("")
	}
	for _, f := range res.Families {
		writeFamily(f)
	}
	for _, f := range res.UnavailableFamilies {
		writeFamily(f)
	}

	// ---------- extra sections ----------
	for _, ex := range meta.Extra {
		w("## %s", ex.Title)
		w("")
		w("%s", strings.TrimRight(ex.Body, "\n"))
		w("")
	}

	// ---------- 7. Caveats ----------
	w("## %d. %s", 7+len(meta.Extra), s.t("局限性与注意事项", "Caveats"))
	w("")
	w("- %s", s.t("延迟为端到端单次小请求的耗时，样本有限，用于横向比较“响应速度”，不是吞吐量或生成速度基准；长输出的 tokens/s 见延迟实验室。", "Latency is the end-to-end time of small single requests with limited samples; use it for relative comparisons, not as a throughput benchmark. See the latency lab for tokens/s."))
	w("- %s", s.t("所有流量经本机网络与可能的代理，绝对数值会随网络、节点与时段变化；同一时段内的相对排序更有参考价值。", "All traffic passes through the local network and possibly proxies; absolute numbers vary with network, nodes and time of day, so relative order within one run is more meaningful."))
	w("- %s", s.t("大量不可用是**提供商账号层面**的问题（无凭证、token 失效、余额/免费额度用尽、套餐限制），这些提供商下的模型本身未必不可用，修复账号后应重新测试。", "Many failures are **provider/account-level** (missing credentials, expired tokens, exhausted balance or quota, plan restrictions); the models themselves may work once the account is fixed."))
	w("- %s", s.t("429 类错误已尽量区分“瞬时限流”与“账号额度用尽”，但依赖报错文本，可能有少量误判；原文均已附在第 5 节。", "429-class errors are split into momentary rate limits and exhausted quotas based on the error text, so a few may be misclassified; the original messages are in section 5."))
	w("- %s", s.t("同一模型系列的归并基于命名规则与从本端点学到的别名，不同提供商对同名模型的实际后端可能不同（见“实际返回模型名”列）。", "Family grouping is rule-based plus aliases learned from this endpoint; different providers may serve different backends under the same name (see the served model column)."))
	return b.String()
}

func dialectNote(meta Meta, s strs) string {
	if meta.Dialect == "" || meta.Dialect == "generic" {
		return ""
	}
	return fmt.Sprintf(s.t("（识别为 %s）", " (detected as %s)"), meta.Dialect)
}

func jitter(x float64) string {
	if x <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f×", x)
}

func flakyNote(m *analyze.ModelStat, s strs) string {
	if len(m.Flaky) == 0 {
		return ""
	}
	var parts []string
	for _, k := range sortedKeys(m.Flaky) {
		parts = append(parts, fmt.Sprintf("%s×%d", reasonName(k, s.lang), m.Flaky[k]))
	}
	return fmt.Sprintf(s.t("曾失败 %s，复测成功", "failed %s, succeeded on retry"), strings.Join(parts, ", "))
}

func httpStr(m *analyze.ModelStat, s strs) string {
	if m.HTTP == 0 {
		return s.t("无响应", "no response")
	}
	if m.Inner > 0 && m.Inner != m.HTTP {
		return fmt.Sprintf("%d/%d", m.HTTP, m.Inner)
	}
	return fmt.Sprintf("%d", m.HTTP)
}
