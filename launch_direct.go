// launch_direct.go
package unrevealed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

type devToolsVersion struct {
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

func newDirect(ctx context.Context, chromePath string, cfg Config) (*Browser, error) {
	port, listener, err := freePort()
	if err != nil {
		return nil, fmt.Errorf("free port: %w", err)
	}

	userDataDir, tmpDir, err := setupDataDir(cfg.UserDataDir)
	if err != nil {
		listener.Close()
		return nil, fmt.Errorf("data dir: %w", err)
	}

	cmd := exec.Command(chromePath, chromeArgs(port, userDataDir, cfg)...)

	listener.Close()
	if err := cmd.Start(); err != nil {
		cleanupTmpDir(tmpDir)
		return nil, fmt.Errorf("start chrome: %w", err)
	}

	b := &Browser{cmd: cmd, tmpDir: tmpDir}

	wsURL, err := resolveWSURL(ctx, port, cfg.ConnectTimeout)
	if err != nil {
		b.Close()
		return nil, fmt.Errorf("connect rod: %w", err)
	}

	browser := rod.New().ControlURL(wsURL).NoDefaultDevice()
	if err := browser.Connect(); err != nil {
		b.Close()
		return nil, fmt.Errorf("connect rod: %w", err)
	}

	b.Browser = browser
	return b, nil
}

func chromeArgs(port int, userDataDir string, cfg Config) []string {
	args := []string{
		"--remote-debugging-host=127.0.0.1",
		fmt.Sprintf("--remote-debugging-port=%d", port),
	}
	return append(args, chromeLaunchArgs(userDataDir, cfg)...)
}

func chromeLaunchArgs(userDataDir string, cfg Config) []string {
	args := []string{
		"--user-data-dir=" + userDataDir,
		"--test-type",
		fmt.Sprintf("--window-size=%d,%d", cfg.WindowWidth, cfg.WindowHeight),
		fmt.Sprintf("--lang=%s", cfg.Language),
		"--log-level=0",
	}

	for flag, val := range StealthFlags() {
		if val != "" {
			args = append(args, fmt.Sprintf("--%s=%s", flag, val))
		} else {
			args = append(args, "--"+flag)
		}
	}

	if cfg.NoSandbox {
		args = append(args, "--no-sandbox")
	}
	if cfg.Headless {
		args = append(args, "--headless=new")
	}

	args = append(args, cfg.ExtraArgs...)
	return args
}

func resolveWSURL(ctx context.Context, port int, timeout time.Duration, baseURLs ...string) (string, error) {
	var versionURL string
	if len(baseURLs) > 0 && baseURLs[0] != "" {
		versionURL = strings.TrimRight(baseURLs[0], "/") + "/json/version"
	} else {
		versionURL = fmt.Sprintf("http://127.0.0.1:%d/json/version", port)
	}

	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		if ws, err := fetchWSURL(ctx, client, versionURL); err == nil && ws != "" {
			return ws, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return "", fmt.Errorf("timeout waiting for chrome debugger at %s", versionURL)
}

func fetchWSURL(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var info devToolsVersion
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	return info.WebSocketDebuggerURL, nil
}
