// tests/unrevealed_test.go
package unrevealed_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/smegg99/unrevealed"
)

func TestPatcherRun(t *testing.T) {
	path, err := unrevealed.FindChrome()
	if err != nil {
		t.Skip("no chrome/chromium found:", err)
	}
	major, err := unrevealed.ChromeMajorVersion(path)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	patcher := unrevealed.NewPatcher(major)
	t.Cleanup(patcher.Cleanup)

	driverPath, err := patcher.Run(ctx)
	if err != nil {
		// ErrCDCNotFound is acceptable the binary may not contain the
		// marker in newer ChromeDriver versions.
		if errors.Is(err, unrevealed.ErrCDCNotFound) {
			t.Skipf("cdc pattern not found (ChromeDriver %d may not contain it)", major)
		}
		t.Fatal(err)
	}

	if driverPath == "" {
		t.Fatal("Run returned empty driver path")
	}
	if patcher.DownloadSHA256 == "" {
		t.Fatal("DownloadSHA256 should be populated after Run")
	}
	if !patcher.IsPatched() {
		t.Fatal("binary should be patched after Run")
	}

	t.Logf("patched chromedriver %d at %s (sha256: %s)", major, driverPath, patcher.DownloadSHA256)
}

func TestNewDirect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	browser, err := unrevealed.New(ctx, unrevealed.Config{
		Headless:  false,
		NoSandbox: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer browser.Close()

	pages := browser.MustPages()
	if len(pages) == 0 {
		t.Fatal("no pages found after launch")
	}
	page := pages[0]

	if err := unrevealed.Stealth(page); err != nil {
		t.Fatal("stealth injection failed:", err)
	}

	page.MustNavigate("https://bot.sannysoft.com/").MustWaitStable()

	result := page.MustEval(`() => navigator.webdriver`)
	if result.String() != "<nil>" {
		t.Errorf("navigator.webdriver should be undefined, got %v", result)
	}
}

func TestNewWithChromeDriver(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	browser, err := unrevealed.New(ctx, unrevealed.Config{
		Headless:          false,
		NoSandbox:         true,
		PatchChromeDriver: true,
	})
	if err != nil {
		if errors.Is(err, unrevealed.ErrCDCNotFound) {
			t.Skip("cdc pattern not found, skipping chromedriver test")
		}
		t.Fatal(err)
	}
	defer browser.Close()

	pages := browser.MustPages()
	if len(pages) == 0 {
		t.Fatal("no pages found after launch")
	}
	page := pages[0]

	if err := unrevealed.Stealth(page); err != nil {
		t.Fatal("stealth injection failed:", err)
	}

	page.MustNavigate("https://bot.sannysoft.com/").MustWaitStable()

	result := page.MustEval(`() => navigator.webdriver`)
	if result.String() != "<nil>" {
		t.Errorf("navigator.webdriver should be undefined, got %v", result)
	}
}
