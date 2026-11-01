package version

import (
	"strings"
	"testing"
)

func TestGetAndString(t *testing.T) {
	info := Get()
	if info.Version != Version || info.Commit != Commit || info.Date != Date {
		t.Fatalf("Get() does not reflect the package variables: %+v", info)
	}
	if info.GoVersion == "" || info.OS == "" || info.Arch == "" {
		t.Fatalf("runtime fields missing: %+v", info)
	}
	s := String()
	for _, want := range []string{"llm-speedtest", Version, Commit, info.OS + "/" + info.Arch} {
		if !strings.Contains(s, want) {
			t.Fatalf("String() = %q, missing %q", s, want)
		}
	}
}
