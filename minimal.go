package unrevealed

import (
	"path"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// blockedResourceTypes are visual/heavy types that the minimalistic
// mode drops to keep raw HTML + JS working.
var blockedResourceTypes = []proto.NetworkResourceType{
	proto.NetworkResourceTypeStylesheet,
	proto.NetworkResourceTypeImage,
	proto.NetworkResourceTypeFont,
	proto.NetworkResourceTypeMedia,
	proto.NetworkResourceTypeManifest,
	proto.NetworkResourceTypeTextTrack,
	proto.NetworkResourceTypePrefetch,
	proto.NetworkResourceTypePing,
}

// Minimal sets up page-level request hijacking that blocks visual/resource-heavy
// CDP types (Stylesheet, Image, Font, Media, Manifest, TextTrack, Prefetch, Ping).
// Call before navigating. For browsers launched with [Config.Minimal], this is
// applied automatically to the default page; call it manually for additional pages.
func (b *Browser) Minimal(page *rod.Page) error {
	router, err := enableMinimalMode(page, b.blockFilenames)
	if err != nil {
		return err
	}
	b.mu.Lock()
	b.hijackRouters = append(b.hijackRouters, router)
	b.mu.Unlock()
	return nil
}

// enableMinimalMode sets up page-level request hijacking that blocks
// resource-heavy types and optionally blocks requests matching
// [Config.BlockFilenames]. The router runs in a background goroutine;
// stop it via the returned router's Stop method or by closing the browser.
func enableMinimalMode(page *rod.Page, blockFilenames []string) (*rod.HijackRouter, error) {
	router := page.HijackRequests()

	lower := make([]string, len(blockFilenames))
	for i, f := range blockFilenames {
		lower[i] = strings.ToLower(f)
	}

	handler := func(h *rod.Hijack) {
		// Block by filename if configured.
		if len(lower) > 0 {
			base := strings.ToLower(path.Base(h.Request.URL().Path))
			for _, name := range lower {
				if base == name {
					h.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
					return
				}
			}
		}

		// Block by resource type.
		rt := h.Request.Type()
		for _, blocked := range blockedResourceTypes {
			if rt == blocked {
				h.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
				return
			}
		}

		h.ContinueRequest(&proto.FetchContinueRequest{})
	}

	if err := router.Add("*", "", handler); err != nil {
		return nil, err
	}

	go router.Run()

	return router, nil
}
