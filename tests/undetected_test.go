// tests/unrevealed_test.go
package unrevealed_test

import (
	"testing"
	"time"

	"github.com/smegg99/unrevealed"
)

func TestBotSannySoft(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	browser, cleanup, err := unrevealed.New(unrevealed.Config{
		Headless: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	// Grab the initial blank page Chrome opened instead of creating a new tab.
	pages := browser.MustPages()
	if len(pages) == 0 {
		t.Fatal("no pages found after launch")
	}
	page := pages[0]

	if err := unrevealed.Stealth(page); err != nil {
		t.Fatal("stealth injection failed:", err)
	}

	page.MustNavigate("https://accounts.censys.com/register/").MustWaitStable()
	time.Sleep(3 * time.Second)
}
