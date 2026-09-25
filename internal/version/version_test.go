package version

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolvePrefersLdflagsThenBuildInfo(t *testing.T) {
	cases := []struct {
		name, ldflags, buildInfo, want string
	}{
		{"ldflags wins", "v1.2.3", "v0.0.1", "v1.2.3"},
		{"build info from go install", "dev", "v0.1.6", "v0.1.6"},
		{"devel build is dev", "dev", "(devel)", "dev"},
		{"nothing known", "dev", "", "dev"},
	}
	for _, c := range cases {
		if got := resolve(c.ldflags, c.buildInfo); got != c.want {
			t.Errorf("%s: resolve(%q, %q) = %q, want %q", c.name, c.ldflags, c.buildInfo, got, c.want)
		}
	}
}

func TestLatestReadsModuleProxy(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"Version":"v0.1.7","Time":"2026-09-25T20:16:55Z"}`))
	}))
	defer srv.Close()

	got, err := Latest(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if got != "v0.1.7" {
		t.Errorf("Latest = %q, want v0.1.7", got)
	}
	if want := "/" + Module + "/@latest"; gotPath != want {
		t.Errorf("queried %q, want %q", gotPath, want)
	}
}

func TestLatestFailsOnProxyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	if _, err := Latest(context.Background(), srv.URL); err == nil {
		t.Fatal("expected an error for a 404 from the proxy")
	}
}

func TestOutdated(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"v0.1.6", "v0.1.6", false},
		{"v0.1.6", "v0.1.7", true},
		{"v0.1.7", "v0.1.6", false}, // proxy lags a fresh tag: never downgrade
		{"v0.1.9", "v0.1.10", true},
		{"v0.1.7+dirty", "v0.1.7", false},
		{"dev", "v0.1.7", true},
	}
	for _, c := range cases {
		if got := Outdated(c.current, c.latest); got != c.want {
			t.Errorf("Outdated(%q, %q) = %v, want %v", c.current, c.latest, got, c.want)
		}
	}
}
