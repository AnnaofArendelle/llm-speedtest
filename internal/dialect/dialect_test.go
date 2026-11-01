package dialect

import (
	"context"
	"testing"

	"github.com/AnnaofArendelle/llm-speedtest/internal/apiclient"
	"github.com/AnnaofArendelle/llm-speedtest/internal/testutil/mockllm"
)

func client(t *testing.T, url string) *apiclient.Client {
	t.Helper()
	c, err := apiclient.New(url+"/v1", "", apiclient.Options{})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestDetect(t *testing.T) {
	srv := mockllm.New(nil)
	defer srv.Close()
	srv.Dialect = "9router"
	srv.Version = "0.5.104"
	info := Detect(context.Background(), client(t, srv.URL))
	if info.Name != NineRouter || info.Version != "0.5.104" || info.Healthy == nil || !*info.Healthy {
		t.Fatalf("9router detect = %+v", info)
	}
	srv.Dialect = "oneapi"
	info = Detect(context.Background(), client(t, srv.URL))
	if info.Name != OneAPI || info.Version != "0.5.104" {
		t.Fatalf("oneapi detect = %+v", info)
	}
	srv.Dialect = ""
	info = Detect(context.Background(), client(t, srv.URL))
	if info.Name != Generic {
		t.Fatalf("generic detect = %+v", info)
	}
	c, _ := apiclient.New("https://openrouter.ai/api/v1", "", apiclient.Options{})
	if Detect(context.Background(), c).Name != OpenRouter {
		t.Fatal("openrouter by host")
	}
	if h := Health(context.Background(), "nope", client(t, srv.URL)); h.Name != Generic {
		t.Fatal("unknown dialect health")
	}
	if len(Names()) != len(All())+1 {
		t.Fatal("names")
	}
}
