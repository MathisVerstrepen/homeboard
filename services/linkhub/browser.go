package linkhub

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func Run(ctx context.Context, url string, verbose bool, timeout time.Duration) (*string, error) {
	fmt.Println("Starting chromedp crawler")
	st := time.Now()

	optsExecAllocator := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck,
		chromedp.ProxyServer("socks5://192.168.2.51:1080"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.3"),
	)

	allocCtx, cancelAllocator := chromedp.NewExecAllocator(ctx, optsExecAllocator...)
	defer cancelAllocator()

	browserCtx, cancelChrome := chromedp.NewContext(allocCtx)
	defer cancelChrome()

	// create chrome instance
	if verbose {
		browserCtx, cancelChrome = chromedp.NewContext(browserCtx, chromedp.WithDebugf(log.Printf))
		defer cancelChrome()
		// No need to defer cancelChrome here again as it's already deferred above and it will reset the cancellation function
	}

	// Create a context with timeout
	browserCtx, cancelTimeout := context.WithTimeout(browserCtx, timeout)
	defer cancelTimeout()

	// capture screenshot
	var buf []byte
	var res string
	if err := chromedp.Run(browserCtx,
		enableLifeCycleEvents(),
		navigateAndWaitFor(url, "networkIdle"),
		chromedp.FullScreenshot(&buf, 90),
		chromedp.ActionFunc(func(c context.Context) error {
			node, err := dom.GetDocument().Do(c)
			if err != nil {
				return err
			}
			res, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(c)
			return err
		}),
	); err != nil {
		return nil, err
	}

	if err := os.WriteFile("fullScreenshot.png", buf, 0o644); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("> Done in %v\n", time.Since(st))

	return &res, nil
}

// This part came from https://github.com/chromedp/chromedp/issues/431#issuecomment-592950397

// enableLifeCycleEvents enables lifecycle events for the browser.
func enableLifeCycleEvents() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		err := page.Enable().Do(ctx)
		if err != nil {
			return err
		}
		err = page.SetLifecycleEventsEnabled(true).Do(ctx)
		if err != nil {
			return err
		}
		return nil
	}
}

// navigateAndWaitFor navigates to the specified URL and waits for the specified event.
// It returns an ActionFunc that can be executed using chromedp.Run.
func navigateAndWaitFor(url string, eventName string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		_, _, _, err := page.Navigate(url).Do(ctx)
		if err != nil {
			return err
		}

		return waitFor(ctx, eventName)
	}
}

// waitFor blocks until eventName is received.
// Examples of events you can wait for:
//
//	init, DOMContentLoaded, firstPaint,
//	firstContentfulPaint, firstImagePaint,
//	firstMeaningfulPaintCandidate,
//	load, networkAlmostIdle, firstMeaningfulPaint, networkIdle
func waitFor(ctx context.Context, eventName string) error {
	ch := make(chan struct{})
	cctx, cancel := context.WithCancel(ctx)
	chromedp.ListenTarget(cctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *page.EventLifecycleEvent:
			if e.Name == eventName {
				cancel()
				close(ch)
			}
		}
	})
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
