// patcher.go
package unrevealed

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

const legacyMaxVersion = 114

const (
	legacyVersionURL  = "https://chromedriver.storage.googleapis.com/LATEST_RELEASE_%d"
	modernVersionURL  = "https://googlechromelabs.github.io/chrome-for-testing/latest-versions-per-milestone-with-downloads.json"
	legacyDownloadURL = "https://chromedriver.storage.googleapis.com/%s/chromedriver_%s.zip"
	modernDownloadURL = "https://storage.googleapis.com/chrome-for-testing-public/%s/%s/chromedriver-%s.zip"
)

const (
	cdcPattern       = `\{window\.cdc.*?;\}`
	patchReplacement = `{console.log("unrevealed chromedriver 1337!")}`
	patchMarker      = "unrevealed chromedriver"
)

const (
	platformWin32    = "win32"
	platformWin64    = "win64"
	platformMac64    = "mac64"
	platformMacArm64 = "mac-arm64"
	platformMacX64   = "mac-x64"
	platformLinux64  = "linux64"

	exeNameWindows = "chromedriver.exe"
	exeNameUnix    = "chromedriver"
)

const dataDirName = "unrevealed"

type milestonesResponse struct {
	Milestones map[string]milestoneInfo `json:"milestones"`
}

type milestoneInfo struct {
	Version string `json:"version"`
}

// Patcher downloads and patches a ChromeDriver binary to remove
// automation detection markers injected by ChromeDriver into the browser.
type Patcher struct {
	MajorVersion int
	DriverPath   string
	DataDir      string
	platform     string
	exeName      string
}

// NewPatcher creates a patcher for the given Chrome major version.
func NewPatcher(majorVersion int) *Patcher {
	p := &Patcher{
		MajorVersion: majorVersion,
		DataDir:      defaultDataDir(),
	}
	p.initPlatform()
	return p
}

// Run downloads the matching ChromeDriver and patches it.
// Returns the path to the patched binary.
func (p *Patcher) Run() (string, error) {
	prepare := func() error {
		if err := os.MkdirAll(p.DataDir, 0o755); err != nil {
			return err
		}
		p.DriverPath = filepath.Join(p.DataDir, randomHex(8)+"_"+p.exeName)
		return nil
	}

	fetchAndDownload := func() ([]byte, error) {
		version, err := p.fetchVersion()
		if err != nil {
			return nil, fmt.Errorf("fetch version: %w", err)
		}
		// slog.Info("resolved chromedriver version", "version", version)
		return p.download(version)
	}

	if err := prepare(); err != nil {
		return "", fmt.Errorf("prepare: %w", err)
	}

	data, err := fetchAndDownload()
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}

	if err := p.extract(data); err != nil {
		return "", fmt.Errorf("extract: %w", err)
	}

	if err := p.patch(); err != nil {
		return "", fmt.Errorf("patch: %w", err)
	}

	// slog.Info("chromedriver ready", "path", p.DriverPath)
	return p.DriverPath, nil
}

// IsPatched reports whether the binary at DriverPath has been patched.
func (p *Patcher) IsPatched() bool {
	data, err := os.ReadFile(p.DriverPath)
	if err != nil {
		return false
	}
	return bytes.Contains(data, []byte(patchMarker))
}

// Cleanup removes the patched binary from disk.
func (p *Patcher) Cleanup() {
	if p.DriverPath != "" {
		os.Remove(p.DriverPath)
	}
}
