# Unrevealed

Undetected Chrome for [go-rod](https://go-rod.github.io/).

Bypasses bot detection by launching Chrome with stealth flags and injecting anti-detection JS via CDP. Passes Cloudflare Turnstile (as of 27.02.2026). Inspired by [carl0smat3us/undetected](https://github.com/carl0smat3us/undetected), rewritten in Go for [ThugHunter](https://github.com/smegg99/ThugHunter).

## Install

```
go get github.com/smegg99/unrevealed
```

### Direct Mode (default)

Chrome is launched directly and controlled through its native DevTools Protocol (CDP) WebSocket. Stealth flags remove automation markers at the Chrome level, and injected JS scripts patch any remaining fingerprinting leaks (`navigator.webdriver`, `window.chrome.runtime`, WebGL, canvas, etc.).

```go
ctx := context.Background()

browser, err := unrevealed.New(ctx, unrevealed.Config{
    Headless: true,
})
if err != nil {
    log.Fatal(err)
}
defer browser.Close()

page := browser.MustPages()[0]
unrevealed.Stealth(page)
page.MustNavigate("https://bot.sannysoft.com/").MustWaitStable()
```

### ChromeDriver Mode

When you need ChromeDriver (e.g., for Selenium-compatible workflows), set `PatchChromeDriver: true`. This automatically downloads a ChromeDriver binary matching your Chrome version, patches it to remove the `cdc_` automation marker that ChromeDriver injects into every page it controls, and uses it to launch Chrme. The patched binary is cleaned up on `browser.Close()`.

```go
ctx := context.Background()

browser, err := unrevealed.New(ctx, unrevealed.Config{
    Headless:          true,
    PatchChromeDriver: true,
})
if err != nil {
    log.Fatal(err)
}
defer browser.Close()

page := browser.MustPages()[0]
unrevealed.Stealth(page)
page.MustNavigate("https://bot.sannysoft.com/").MustWaitStable()
```

You can also supply a pre-patched binary directly:

```go
browser, err := unrevealed.New(ctx, unrevealed.Config{
    ChromeDriverPath: "/path/to/patched/chromedriver",
})
```

### Virtual Display (Xvfb)

On Linux, set `VirtualDisplay: true` to run Chrome in a virtual X11 display via Xvfb. This gives you a full headed browser environment without needing a physical display, some bot protections can be avoided this way. Headless mode is automatically disabled when using a virtual display.

```go
browser, err := unrevealed.New(ctx, unrevealed.Config{
    VirtualDisplay: true,
    NoSandbox:      true,
})
```

Requires `Xvfb` to be installed.

### Minimal Mode

Set `Minimal: true` to block visual/resource-heavy requests (CSS, images, fonts, media, manifests, text tracks, prefetch, ping) browser-wide via CDP request hijacking. HTML and JS keep working normally. Might trigger bot detections that check for missing resources, but can speed up page loads and reduce memory usage when you only care about HTML/JS.

```go
browser, err := unrevealed.New(ctx, unrevealed.Config{
    Headless: true,
    Minimal:  true,
})
```

You can also block specific files by name:

```go
browser, err := unrevealed.New(ctx, unrevealed.Config{
    Headless:       true,
    Minimal:        true,
    BlockFilenames: []string{"analytics.js", "pendo.js"},
})
```

### Browser Memory

`BrowserMemory()` returns the aggregated RSS and VMS of the entire browser process tree (root + all child processes), using the OS process table — not just Go runtime memory.

```go
stats, err := browser.BrowserMemory()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Browser RSS: %d MB\n", stats.RSS/1024/1024)
```

## License

MIT
