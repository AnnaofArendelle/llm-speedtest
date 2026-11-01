package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AnnaofArendelle/llm-speedtest/internal/capability"
	"github.com/AnnaofArendelle/llm-speedtest/internal/latency"
	"github.com/AnnaofArendelle/llm-speedtest/internal/load"
	"github.com/AnnaofArendelle/llm-speedtest/internal/quality"
)

func ms(x float64) string {
	if x <= 0 {
		return "-"
	}
	if x < 1 {
		return fmt.Sprintf("%.0f ms", x*1000)
	}
	return fmt.Sprintf("%.2f s", x)
}

// LatencySection renders latency lab statistics.
func LatencySection(stats []latency.ModelStats, lang string) Section {
	s := strs{lang: lang}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", s.t("流式请求的首 token 时间（TTFT）、首二 token 间隔（TTST）、每输出 token 时间（TPOT）、端到端延迟（E2E）与解码吞吐（tok/s，N−1 定义）。统计不含预热样本；“冷”列为首个预热样本。", "Streaming time-to-first-token (TTFT), time-to-second-token (TTST), time per output token (TPOT), end-to-end latency (E2E) and decode throughput (tok/s, N−1 definition). Statistics exclude warm-up samples; the cold column is the first warm-up sample."))
	fmt.Fprintf(&b, "| # | %s | %s | TTFT p50 | TTFT p90 | TTST p50 | TPOT p50 | E2E p50 | tok/s p50 | tok/s p90 | %s | %s | %s |\n", s.t("模型", "Model"), s.t("样本(成功)", "Samples (ok)"), s.t("冷启动 TTFT", "Cold TTFT"), s.t("输出 token", "Output tokens"), s.t("备注", "Notes"))
	b.WriteString("|---|---|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, m := range stats {
		note := ""
		if !m.TokensExact && m.OK > 0 {
			note = s.t("token 数为估算", "token counts estimated")
		}
		if len(m.Reasons) > 0 {
			var parts []string
			for _, k := range sortedKeys(m.Reasons) {
				parts = append(parts, fmt.Sprintf("%s×%d", reasonName(k, lang), m.Reasons[k]))
			}
			note = strings.TrimSpace(note + " " + strings.Join(parts, ", "))
		}
		cold := "-"
		if m.Cold != nil {
			cold = ms(m.Cold.TTFT)
		}
		rank := "-"
		if m.Rank > 0 {
			rank = fmt.Sprint(m.Rank)
		}
		fmt.Fprintf(&b, "| %s | %s | %d(%d) | %s | %s | %s | %s | %s | %.1f | %.1f | %s | %.0f | %s |\n", rank, code(m.ModelID), m.Samples, m.OK, ms(m.TTFT.Median), ms(m.TTFT.P90), ms(m.TTST.Median), ms(m.TPOT.Median), ms(m.E2E.Median), m.TokPerSec.Median, m.TokPerSec.P90, cold, m.OutputTokens.Median, esc(note))
	}
	return Section{Title: s.t("延迟实验室", "Latency lab"), Body: b.String()}
}

// LoadSection renders load test levels.
func LoadSection(model string, levels []load.LevelResult, lang string) Section {
	s := strs{lang: lang}
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s\n\n", s.t("模型：", "Model:"), code(model))
	fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s | TTFT p50 | TTFT p99 | E2E p50 | E2E p99 | ITL p50 | %s |\n", s.t("级别", "Level"), s.t("模式", "Mode"), s.t("请求数", "Requests"), s.t("成功", "OK"), s.t("错误", "Errors"), s.t("请求吞吐 req/s", "Req/s"), s.t("输出吞吐 tok/s", "Out tok/s"), s.t("goodput", "Goodput"))
	b.WriteString("|---|---|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, l := range levels {
		lvl := fmt.Sprint(l.Level)
		if l.Mode == load.ModeRate {
			lvl = fmt.Sprintf("%.2f req/s", float64(l.Level)/100)
		}
		fmt.Fprintf(&b, "| %s | %s | %d | %d | %d | %.2f | %.1f | %s | %s | %s | %s | %s | %.0f%% |\n", lvl, l.Mode, l.Requests, l.OK, l.Errors, l.RPS, l.OutTPS, ms(l.TTFT.Median), ms(l.TTFT.P99), ms(l.E2E.Median), ms(l.E2E.P99), ms(l.ITL.Median), 100*l.Goodput)
	}
	return Section{Title: s.t("负载测试", "Load test"), Body: b.String()}
}

// CapabilitySection renders the capability matrix.
func CapabilitySection(mx capability.Matrix, lang string) Section {
	s := strs{lang: lang}
	var b strings.Builder
	b.WriteString("| " + s.t("模型", "Model") + " | " + s.t("支持数", "Supported"))
	for _, p := range mx.Probes {
		b.WriteString(" | " + capability.Describe(p, lang))
	}
	b.WriteString(" |\n|---|---")
	for range mx.Probes {
		b.WriteString("|---")
	}
	b.WriteString("|\n")
	for _, m := range mx.Models {
		fmt.Fprintf(&b, "| %s | %d/%d", code(m), mx.Support[m], len(mx.Probes))
		for _, p := range mx.Probes {
			c := mx.Cells[m][p]
			sym := "·"
			switch c.Status {
			case capability.Supported:
				sym = "✅"
			case capability.Unsupported:
				sym = "❌"
			case capability.Error:
				sym = "⚠️"
			}
			b.WriteString(" | " + sym)
		}
		b.WriteString(" |\n")
	}
	b.WriteString("\n" + s.t("证据明细：", "Evidence:") + "\n\n")
	for _, m := range mx.Models {
		fmt.Fprintf(&b, "- %s\n", code(m))
		for _, p := range mx.Probes {
			c := mx.Cells[m][p]
			if c.Evidence != "" {
				fmt.Fprintf(&b, "  - %s: %s — %s\n", capability.Describe(p, lang), c.Status, esc(trunc(c.Evidence, 140)))
			}
		}
	}
	return Section{Title: s.t("能力矩阵", "Capability matrix"), Body: b.String()}
}

// QualitySection renders quality scores.
func QualitySection(scores []quality.ModelScore, lang string) Section {
	s := strs{lang: lang}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", s.t("内置确定性题集（无需裁判模型），分数为各次试验加权通过率的中位数；一致性为多次试验结果一致的题目占比；“自述身份”为模型对“你是什么模型”的回答，仅供参考。", "Built-in deterministic question set (no judge model). Score is the median over trials of the weighted pass rate; consistency is the share of questions answered identically across trials; identity is the model's own answer to “which model are you”."))
	fmt.Fprintf(&b, "| # | %s | %s | %s | %s | %s | %s |\n", s.t("模型", "Model"), s.t("质量分", "Score"), s.t("试验", "Trials"), s.t("一致性", "Consistency"), s.t("自述身份", "Identity"), s.t("分类通过率", "Category pass rates"))
	b.WriteString("|---|---|---|---|---|---|---|\n")
	for _, m := range scores {
		var cats []string
		keys := make([]string, 0, len(m.Categories))
		for k := range m.Categories {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			cats = append(cats, fmt.Sprintf("%s %.0f%%", k, 100*m.Categories[k]))
		}
		rank := "-"
		if m.Rank > 0 {
			rank = fmt.Sprint(m.Rank)
		}
		score := "-"
		if m.Answered > 0 {
			score = fmt.Sprintf("%.1f", m.Score)
		}
		errs := ""
		if len(m.Errors) > 0 {
			errs = " " + joinCounts(m.Errors, lang, 2, true)
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %d | %.0f%% | %s | %s%s |\n", rank, code(m.ModelID), score, m.Trials, 100*m.Consistency, trunc(m.Identity, 40), esc(strings.Join(cats, ", ")), esc(errs))
	}
	return Section{Title: s.t("质量抽检", "Quality spot-check"), Body: b.String()}
}
