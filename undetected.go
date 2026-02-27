// undetected.go
package unrevealed

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// Config controls the behavior of the undetected browser.
type Config struct {
	// Chrome executable path. Auto-detected if empty.
	ChromePath string

	// Run Chrome in headless mode.
	Headless bool

	// User data directory for the Chrome profile. Uses a temp dir if empty.
	UserDataDir string

	// Additional raw Chrome arguments.
	ExtraArgs []string
}

// New launches a stealth-configured Chrome browser and connects go-rod to it.
func New(cfg Config) (*rod.Browser, func(), error) {
	resolvePath := func() (string, error) {
		if cfg.ChromePath != "" {
			return cfg.ChromePath, nil
		}
		return FindChrome()
	}

	setupDataDir := func() (userDataDir, tmpDir string, err error) {
		if cfg.UserDataDir != "" {
			return cfg.UserDataDir, "", nil
		}
		tmp, err := os.MkdirTemp("", "unrevealed-*")
		if err != nil {
			return "", "", err
		}
		return tmp, tmp, nil
	}

	startChrome := func(chromePath string, port int, userDataDir string) (*exec.Cmd, error) {
		cmd := exec.Command(chromePath, chromeArgs(port, userDataDir, cfg)...)
		return cmd, cmd.Start()
	}

	connectRod := func(port int) (*rod.Browser, error) {
		wsURL, err := resolveWSURL(port, 15*time.Second)
		if err != nil {
			return nil, err
		}
		slog.Info("chrome ready", "ws", wsURL)
		b := rod.New().ControlURL(wsURL).NoDefaultDevice()
		return b, b.Connect()
	}

	chromePath, err := resolvePath()
	if err != nil {
		return nil, nil, fmt.Errorf("find chrome: %w", err)
	}
	slog.Info("using chrome", "path", chromePath)

	port, err := freePort()
	if err != nil {
		return nil, nil, fmt.Errorf("free port: %w", err)
	}

	userDataDir, tmpDir, err := setupDataDir()
	if err != nil {
		return nil, nil, fmt.Errorf("data dir: %w", err)
	}

	cmd, err := startChrome(chromePath, port, userDataDir)
	if err != nil {
		cleanupTmpDir(tmpDir)
		return nil, nil, fmt.Errorf("start chrome: %w", err)
	}

	killChrome := func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		cleanupTmpDir(tmpDir)
	}

	browser, err := connectRod(port)
	if err != nil {
		killChrome()
		return nil, nil, fmt.Errorf("connect rod: %w", err)
	}

	cleanup := func() {
		_ = browser.Close()
		killChrome()
	}

	return browser, cleanup, nil
}

// Stealth injects anti-detection scripts into a rod page.
// Call before navigating to the target URL. Scripts persist across navigations.
func Stealth(page *rod.Page) error {
	_ = proto.EmulationClearDeviceMetricsOverride{}.Call(page)

	for _, script := range StealthScripts() {
		_, err := proto.PageAddScriptToEvaluateOnNewDocument{
			Source: script,
		}.Call(page)
		if err != nil {
			return fmt.Errorf("inject stealth script: %w", err)
		}
	}
	return nil
}

func chromeArgs(port int, userDataDir string, cfg Config) []string {
	args := []string{
		"--remote-debugging-host=127.0.0.1",
		fmt.Sprintf("--remote-debugging-port=%d", port),
		"--user-data-dir=" + userDataDir,
		"--disable-blink-features=AutomationControlled",
		"--no-first-run",
		"--no-default-browser-check",
		"--no-sandbox",
		"--test-type",
		"--window-size=1920,1080",
		"--start-maximized",
		"--lang=en-US",
		"--log-level=0",
	}

	if cfg.Headless {
		args = append(args, "--headless=new")
	}

	args = append(args, cfg.ExtraArgs...)
	return args
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func resolveWSURL(port int, timeout time.Duration) (string, error) {
	url := fmt.Sprintf("http://127.0.0.1:%d/json/version", port)
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		if ws, err := fetchWSURL(client, url); err == nil && ws != "" {
			return ws, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return "", fmt.Errorf("timeout waiting for chrome on port %d", port)
}

func fetchWSURL(client *http.Client, url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var info struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	return info.WebSocketDebuggerURL, nil
}

func cleanupTmpDir(dir string) {
	if dir != "" {
		os.RemoveAll(dir)
	}
}
