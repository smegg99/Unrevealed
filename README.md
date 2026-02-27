# Unrevealed

Undetected Chrome for [go-rod](https://go-rod.github.io/).

Bypasses bot detection by launching Chrome directly with minimal flags instead of go-rod's launcher (which adds a bunch of detectable automation flags), injecting stealth JS via CDP, and optionally patching ChromeDriver to remove the `cdc_` marker. Passes Cloudflare Turnstile (as of 27.02.2026). Inspired by [carl0smat3us/undetected](https://github.com/carl0smat3us/undetected), rewritten in Go for [ThugHunter](https://github.com/smegg99/ThugHunter).

## Install

```
go get github.com/smegg99/unrevealed
```

## Usage

```go
browser, cleanup, err := unrevealed.New(unrevealed.Config{
    ChromePath:  "/usr/bin/google-chrome", // auto-detected if empty
    Headless:    true,
    UserDataDir: "/tmp/my-profile",        // temp dir if empty
    ExtraArgs:   []string{"--proxy-server=socks5://127.0.0.1:1080"}, // you can add any additional chrome launch flags here
})
if err != nil {
    log.Fatal(err)
}
defer cleanup()

page := browser.MustPages()[0] // reuse Chrome's initial tab
unrevealed.Stealth(page)

page.MustNavigate("https://bot.sannysoft.com/").MustWaitStable()
```

Stealth scripts and flags are also exported individually for use with chromedp or other CDP libraries (see `StealthScripts()`, `StealthFlags()`, `DeleteFlags()`).

## License

MIT
