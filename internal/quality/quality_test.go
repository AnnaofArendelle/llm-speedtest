package quality

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/apiclient"
)

func TestDefaultSetAndChecks(t *testing.T) {
	set, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Questions) != 20 || set.IdentityPrompt == "" {
		t.Fatalf("set = %d questions", len(set.Questions))
	}
	byID := map[string]*Question{}
	for i := range set.Questions {
		byID[set.Questions[i].ID] = &set.Questions[i]
	}
	cases := []struct {
		q, out string
		pass   bool
	}{
		{"multiply", "391", true}, {"multiply", "The answer is 391.", true}, {"multiply", "17 × 23 = 391", true}, {"multiply", "390", false},
		{"number_words", "423", true}, {"number_words", "4 2 3", false}, {"number_words", "1,024", false},
		{"exact_word", "APPLE", true}, {"exact_word", "apple.", true}, {"exact_word", "APPLE PIE", false},
		{"reverse", "kramhcneb", true}, {"reverse", "KRAMHCNEB", true}, {"reverse", "benchmark", false},
		{"json_extract", `{"city": "Oslo"}`, true}, {"json_extract", `{"city":"Lisbon"}`, false},
		{"syllogism", "Yes.", true}, {"syllogism", "No, not necessarily", false},
		{"sort_numbers", "3,7,19,42", true}, {"sort_numbers", "3, 7, 19, 42", true}, {"sort_numbers", "42,19,7,3", false},
		{"translate", "Thank you", true}, {"capital", "Canberra", true}, {"capital", "Sydney", false},
	}
	for _, c := range cases {
		if got := byID[c.q].Check(c.out); got != c.pass {
			t.Errorf("%s %q -> %v, want %v", c.q, c.out, got, c.pass)
		}
	}
	if _, err := Parse([]byte("questions: []")); err == nil {
		t.Fatal("empty set must fail")
	}
	if _, err := Parse([]byte("questions:\n  - id: a\n    prompt: p\n    type: regex\n    expect: '('\n")); err == nil {
		t.Fatal("bad regex must fail")
	}
	if _, err := Parse([]byte("questions:\n  - id: a\n    prompt: p\n    type: nope\n    expect: x\n")); err == nil {
		t.Fatal("bad type must fail")
	}
}

// answerServer returns a model that answers correctly for a subset of
// question ids and nonsense otherwise.
func answerServer(t *testing.T, correct map[string]string, identity string, flakyQ string) *httptest.Server {
	t.Helper()
	calls := map[string]int{}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Model    string `json:"model"`
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Model == "bad" {
			w.WriteHeader(404)
			fmt.Fprint(w, `{"error":{"message":"The model 'bad' does not exist"}}`)
			return
		}
		prompt := body.Messages[0].Content
		answer := "I do not know"
		if strings.Contains(prompt, "Which AI model are you") {
			answer = identity
		}
		for key, ans := range correct {
			if strings.Contains(prompt, key) {
				answer = ans
			}
		}
		if flakyQ != "" && strings.Contains(prompt, flakyQ) {
			calls[flakyQ]++
			if calls[flakyQ]%2 == 0 {
				answer = "wrong"
			}
		}
		fmt.Fprintf(w, `{"model":"%s-served","choices":[{"message":{"content":%q},"finish_reason":"stop"}],"usage":{"completion_tokens":2}}`, body.Model, answer)
	}))
}

func TestRunAndSummarize(t *testing.T) {
	correct := map[string]string{"17 × 23": "391", "four hundred": "423", "APPLE": "APPLE", "Reverse the string": "kramhcneb", "sum of all integers": "55"}
	srv := answerServer(t, correct, "gpt-test-2", "17 × 23")
	defer srv.Close()
	cl, _ := apiclient.New(srv.URL+"/v1", "", apiclient.Options{})
	r := &Runner{Client: cl, Spec: Spec{Trials: 2, Timeout: 5 * time.Second}}
	results, err := r.Run(context.Background(), []Target{{ID: "good", Provider: "p"}, {ID: "bad", Provider: "p"}})
	if err != nil {
		t.Fatal(err)
	}
	scores := Summarize(results, nil)
	if len(scores) != 2 || scores[0].ModelID != "good" || scores[1].ModelID != "bad" {
		t.Fatalf("scores = %+v", scores)
	}
	good := scores[0]
	// 5 correct of 20 in trial 1; multiply flips in trial 2 -> 4 of 20
	if good.Trials != 2 || len(good.TrialScores) != 2 || good.TrialScores[0] != 25 || good.TrialScores[1] != 20 || good.Score != 22.5 {
		t.Fatalf("good = %+v", good)
	}
	if good.Identity != "gpt-test-2" || good.Rank != 1 || good.Consistency != 0.95 || good.Questions["multiply"] != 0.5 || good.Categories["instruction"] != 1 {
		t.Fatalf("good details = %+v", good)
	}
	bad := scores[1]
	if bad.Answered != 0 || bad.Errors["model_unavailable"] == 0 || bad.Rank != 0 {
		t.Fatalf("bad = %+v", bad)
	}
	// identity failure stops early: 1 result only
	n := 0
	for _, res := range results {
		if res.ModelID == "bad" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("bad results = %d", n)
	}
}
