// Package version reports the running url2svg version and updates the
// installed binary from the Go module proxy.
package version

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"
	"time"
)

// Module is the Go module path url2svg is installed from.
const Module = "git.bytestone.uk/hum3/url2svg"

// DefaultProxy is the Go module proxy queried for the latest release.
const DefaultProxy = "https://proxy.golang.org"

// ldflagsVersion is set by goreleaser via -ldflags "-X ...version.ldflagsVersion=v1.2.3".
// A plain `go install` leaves it as "dev" and the module version is read from build info.
var ldflagsVersion = "dev"

// Current returns the version of the running binary, or "dev" when unknown.
func Current() string {
	buildInfo := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		buildInfo = info.Main.Version
	}
	return resolve(ldflagsVersion, buildInfo)
}

func resolve(ldflags, buildInfo string) string {
	if ldflags != "dev" {
		return ldflags
	}
	if buildInfo != "" && buildInfo != "(devel)" {
		return buildInfo
	}
	return "dev"
}

// Latest asks the module proxy for the newest tagged version of Module.
func Latest(ctx context.Context, proxy string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, proxy+"/"+Module+"/@latest", nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("module proxy returned %s", resp.Status)
	}
	var latest struct{ Version string }
	if err := json.NewDecoder(resp.Body).Decode(&latest); err != nil {
		return "", fmt.Errorf("decoding proxy response: %w", err)
	}
	if latest.Version == "" {
		return "", fmt.Errorf("module proxy response had no version")
	}
	return latest.Version, nil
}

// Outdated reports whether current should be replaced by latest. A "dev"
// build is always considered outdated so `self update` installs a release.
func Outdated(current, latest string) bool {
	return current != latest
}

// Update installs the latest release over the running binary, reporting
// progress to w. It is a no-op when the current version is already latest.
func Update(ctx context.Context, w io.Writer) error {
	current := Current()
	_, _ = fmt.Fprintf(w, "Current version: %s\n", current)

	latest, err := Latest(ctx, DefaultProxy)
	if err != nil {
		return fmt.Errorf("checking latest version: %w", err)
	}
	_, _ = fmt.Fprintf(w, "Latest version:  %s\n", latest)
	if !Outdated(current, latest) {
		_, _ = fmt.Fprintln(w, "Already up to date.")
		return nil
	}

	_, _ = fmt.Fprintf(w, "Installing: go install %s/cmd/url2svg@%s\n", Module, latest)
	cmd := exec.CommandContext(ctx, "go", "install", Module+"/cmd/url2svg@"+latest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go install: %w", err)
	}
	_, _ = fmt.Fprintln(w, "Updated successfully.")
	return nil
}
