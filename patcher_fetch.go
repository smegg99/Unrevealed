// patcher_fetch.go
package unrevealed

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (p *Patcher) fetchVersion() (string, error) {
	if p.MajorVersion <= legacyMaxVersion {
		return p.fetchLegacyVersion()
	}
	return p.fetchModernVersion()
}

func (p *Patcher) fetchLegacyVersion() (string, error) {
	url := fmt.Sprintf(legacyVersionURL, p.MajorVersion)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

func (p *Patcher) fetchModernVersion() (string, error) {
	resp, err := http.Get(modernVersionURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result milestonesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	ms, ok := result.Milestones[fmt.Sprintf("%d", p.MajorVersion)]
	if !ok {
		return "", fmt.Errorf("chromedriver milestone %d not found", p.MajorVersion)
	}
	return ms.Version, nil
}

func (p *Patcher) download(version string) ([]byte, error) {
	url := p.buildDownloadURL(version)

	// slog.Info("downloading chromedriver", "url", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: HTTP %d from %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

func (p *Patcher) buildDownloadURL(version string) string {
	if p.MajorVersion <= legacyMaxVersion {
		return fmt.Sprintf(legacyDownloadURL, version, p.platform)
	}
	return fmt.Sprintf(modernDownloadURL, version, p.platform, p.platform)
}
