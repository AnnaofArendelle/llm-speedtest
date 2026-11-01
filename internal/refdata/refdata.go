// Package refdata embeds the recorded reference scan (a real multi-provider
// gateway probed on 2026-11-01, 2306 models, 2939 attempts). It powers the
// demo mode and the regression tests.
package refdata

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"embed"
	"encoding/json"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/classify"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/modelid"
)

//go:embed data/*
var data embed.FS

// ListItem is one discovered model of the reference endpoint.
type ListItem struct {
	ID      string `json:"id"`
	OwnedBy string `json:"owned_by"`
}

type rawAttempt struct {
	ID               string  `json:"id"`
	OwnedBy          string  `json:"owned_by"`
	Pass             int     `json:"pass"`
	TS               float64 `json:"ts"`
	Status           int     `json:"status"`
	InnerStatus      int     `json:"inner_status"`
	Conn             string  `json:"conn"`
	Latency          float64 `json:"latency"`
	OK               bool    `json:"ok"`
	Msg              string  `json:"msg"`
	ServedModel      string  `json:"served_model"`
	FinishReason     string  `json:"finish_reason"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
}

// List returns the reference model list.
func List() ([]ListItem, error) {
	raw, err := data.ReadFile("data/testlist.json")
	if err != nil {
		return nil, err
	}
	var items []ListItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// Load returns the reference models and attempts, re-classifying every
// failure with the current classifier.
func Load() ([]core.RunModel, []core.Attempt, error) {
	items, err := List()
	if err != nil {
		return nil, nil, err
	}
	models := make([]core.RunModel, 0, len(items))
	for _, it := range items {
		models = append(models, core.RunModel{ID: it.ID, OwnedBy: it.OwnedBy, Upstream: modelid.Upstream(it.ID, it.OwnedBy)})
	}
	raw, err := data.ReadFile("data/scan_9router_20261101.jsonl.gz")
	if err != nil {
		return nil, nil, err
	}
	gz, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, nil, err
	}
	sc := bufio.NewScanner(gz)
	sc.Buffer(make([]byte, 1024*1024), 4*1024*1024)
	var attempts []core.Attempt
	for sc.Scan() {
		var r rawAttempt
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			return nil, nil, err
		}
		a := core.Attempt{ModelID: r.ID, Round: r.Pass, At: time.Unix(0, int64(r.TS*1e9)), HTTP: r.Status,
			Latency: time.Duration(r.Latency * float64(time.Second)), ServedModel: r.ServedModel, FinishReason: r.FinishReason,
			PromptTokens: r.PromptTokens, CompletionTokens: r.CompletionTokens}
		if r.OK {
			a.OK, a.Class, a.Reason = true, "ok", "ok"
		} else {
			name := modelid.Name(r.ID, r.OwnedBy)
			human, code, conn := classify.DeepUnwrap(r.Msg, r.InnerStatus)
			cls := classify.Class(r.Status, code, human)
			if r.Status == 200 && (cls == "bad_request" || cls == "unknown" || cls == "http_200") {
				cls = "error_in_200"
			}
			a.Class, a.Reason, a.Inner, a.Message = cls, classify.Refine(cls, human, name), code, classify.CleanMessage(human)
			a.Conn = r.Conn
			if a.Conn == "" {
				a.Conn = conn
			}
		}
		attempts = append(attempts, a)
	}
	return models, attempts, sc.Err()
}
