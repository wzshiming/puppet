package puppet

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
	"unsafe"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/client"
	"github.com/chromedp/chromedp/runner"
)

// Puppet DevTools Protocol browser manager, handling the
type Puppet struct {
	cdp    *chromedp.CDP
	cli    *client.Client
	ctx    context.Context
	cancel func()
}

// NewPuppet creates and starts a new CDP instance
func NewPuppet(url string) (*Puppet, error) {

	p := &Puppet{}

	p.ctx, p.cancel = context.WithCancel(context.Background())

	if url == "" {
		listen, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", 9222))
		if err == nil {
			listen.Close()

			run, err := runner.New()
			if err != nil {
				return nil, err
			}
			p.cli = run.Client()

			err = run.Start(p.ctx)
			if err != nil {
				return nil, err
			}
			cdp, err := chromedp.New(p.ctx,
				chromedp.WithRunner(run),
			)
			if err != nil {
				return nil, err
			}
			p.cdp = cdp
			return p, nil
		}
		url = client.DefaultEndpoint
	}

	p.cli = client.New(client.URL(url))
	cdp, err := chromedp.New(p.ctx,
		//	chromedp.WithLog(log.Printf),
		chromedp.WithClient(p.ctx, p.cli),
	)
	if err != nil {
		return nil, err
	}
	p.cdp = cdp

	return p, nil
}

// Close closes all Puppet page handlers.
func (c *Puppet) Close() error {
	c.cancel()
	// shutdown chrome
	err := c.cdp.Shutdown(c.ctx)
	if err != nil {
		return err
	}
	// wait for chrome to finish
	err = c.cdp.Wait()
	if err != nil {
		return err
	}
	return nil
}

// NewTarget an action that creates a new Chrome target, and sets it as the active target.
func (c *Puppet) NewTarget(url string) (id string, err error) {
	t, err := c.cli.NewPageTargetWithURL(c.ctx, url)
	if err != nil {
		return "", err
	}
	return t.GetID(), nil
}

// CloseTarget closes the Chrome target with the specified id.
func (c *Puppet) CloseTarget(id string) (err error) {
	return c.cdp.Run(c.ctx,
		c.cdp.CloseByID(id))
}

// SetTarget is an action that sets the active Chrome handler to the handler associated with the specified id.
func (c *Puppet) SetTarget(id string) (err error) {
	return c.cdp.Run(c.ctx,
		c.cdp.SetTargetByID(id))
}

// Targets returns the target IDs of the managed targets.
func (c *Puppet) Targets() (tabs []string, err error) {
	return c.cdp.ListTargets(), nil
}

// Navigate navigates the current frame.
func (c *Puppet) Navigate(url string) error {
	return c.cdp.Run(c.ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		waitComplete,
	})
}

// NavigateBack navigates the current frame backwards in its history.
func (c *Puppet) NavigateBack() error {
	return c.cdp.Run(c.ctx, chromedp.Tasks{
		chromedp.NavigateBack(),
		waitComplete,
	})
}

// NavigateForward navigates the current frame forwards in its history.
func (c *Puppet) NavigateForward() error {
	return c.cdp.Run(c.ctx, chromedp.Tasks{
		chromedp.NavigateForward(),
		waitComplete,
	})
}

// Reload reloads the current page.
func (c *Puppet) Reload() error {
	return c.cdp.Run(c.ctx, chromedp.Tasks{
		chromedp.Reload(),
		waitComplete,
	})
}

// Stop stops all navigation and pending resource retrieval.
func (c *Puppet) Stop() error {
	return c.cdp.Run(c.ctx,
		chromedp.Stop(),
	)
}

// WaitReady waits until the element is ready (ie, loaded by chromedp).
func (c *Puppet) WaitReady(sel string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.WaitReady(sel))
}

// WaitVisible waits until the selected element is visible.
func (c *Puppet) WaitVisible(sel string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.WaitVisible(sel))
}

// WaitNotVisible waits until the selected element is not visible.
func (c *Puppet) WaitNotVisible(sel string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.WaitNotVisible(sel))
}

// WaitEnabled waits until the selected element is enabled (does not have attribute 'disabled').
func (c *Puppet) WaitEnabled(sel string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.WaitEnabled(sel))
}

// WaitSelected waits until the element is selected (has attribute 'selected').
func (c *Puppet) WaitSelected(sel string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.WaitSelected(sel))
}

// WaitNotPresent waits until no elements match the specified selector.
func (c *Puppet) WaitNotPresent(sel string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.WaitNotPresent(sel))
}

// Evaluate is an action to evaluate the Javascript expression, unmarshaling the result of the script evaluation to res.
func (c *Puppet) Evaluate(expression string, res interface{}) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.Evaluate(expression, res))
}

// Location retrieves the document location.
func (c *Puppet) Location() (url string, err error) {
	return url, c.cdp.Run(c.ctx,
		chromedp.Location(&url))
}

// Title retrieves the document title.
func (c *Puppet) Title() (title string, err error) {
	return title, c.cdp.Run(c.ctx,
		chromedp.Title(&title))
}

// Click sends a mouse click event to the first node matching the selector.
func (c *Puppet) Click(sel string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.Click(sel, chromedp.NodeVisible))
}

// DoubleClick sends a mouse double click event to the first node matching the selector.
func (c *Puppet) DoubleClick(sel string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.DoubleClick(sel, chromedp.NodeVisible))
}

// OuterHTML retrieves the outer html of the first node matching the selector.
func (c *Puppet) OuterHTML() (res []byte, err error) {
	var src string
	err = c.cdp.Run(c.ctx,
		chromedp.OuterHTML("html", &src, chromedp.ByQuery),
	)
	if err != nil {
		return nil, err
	}
	return *(*[]byte)(unsafe.Pointer(&src)), nil
}

// InnerHTML retrieves the inner html of the first node matching the selector.
func (c *Puppet) InnerHTML() (res []byte, err error) {
	var src string
	err = c.cdp.Run(c.ctx,
		chromedp.InnerHTML("html", &src, chromedp.ByQuery),
	)
	if err != nil {
		return nil, err
	}
	return *(*[]byte)(unsafe.Pointer(&src)), nil
}

// SetValue sets the value of an element.
func (c *Puppet) SetValue(sel string, value string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.SetValue(sel, value))
}

// Value retrieves the value of the first node matching the selector.
func (c *Puppet) Value(sel string) (value string, err error) {
	return value, c.cdp.Run(c.ctx,
		chromedp.Value(sel, &value))
}

// Text retrieves the visible text of the first node matching the selector.
func (c *Puppet) Text(sel string) (value string, err error) {
	return value, c.cdp.Run(c.ctx,
		chromedp.Text(sel, &value))
}

// Clear clears the values of any input/textarea nodes matching the selector.
func (c *Puppet) Clear(sel string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.Clear(sel))
}

// Focus focuses the first node matching the selector.
func (c *Puppet) Focus(sel string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.Focus(sel))
}

// KeyAction will synthesize a keyDown, char, and keyUp event for each rune contained in keys along with any supplied key options.
func (c *Puppet) KeyAction(key string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.KeyAction(key))
}

// SetAttributes sets the element attributes for the first node matching the selector.
func (c *Puppet) SetAttributes(sel string, value map[string]string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.SetAttributes(sel, value))
}

// Attributes retrieves the element attributes for the first node matching the selector.
func (c *Puppet) Attributes(sel string) (value map[string]string, err error) {
	return value, c.cdp.Run(c.ctx,
		chromedp.Attributes(sel, &value))
}

// AttributesAll retrieves the element attributes for all nodes matching the selector.
func (c *Puppet) AttributesAll(sel string) (value []map[string]string, err error) {
	return value, c.cdp.Run(c.ctx,
		chromedp.AttributesAll(sel, &value))
}

// SetAttributeValue sets the element attribute with name to value for the first node matching the selector.
func (c *Puppet) SetAttributeValue(sel string, name, value string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.SetAttributeValue(sel, name, value))
}

// AttributeValue retrieves the element attribute value for the first node matching the selector.
func (c *Puppet) AttributeValue(sel string, name string) (value string, ok bool, err error) {
	return value, ok, c.cdp.Run(c.ctx,
		chromedp.AttributeValue(sel, name, &value, &ok))
}

// DelAttribute removes the element attribute with name from the first node matching the selector.
func (c *Puppet) DelAttribute(sel string, name string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.RemoveAttribute(sel, name))
}

// SendKeys synthesizes the key up, char, and down events as needed for the runes in v, sending them to the first node matching the selector.
func (c *Puppet) SendKeys(sel string, v string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.SendKeys(sel, v))
}

// Submit is an action that submits the form of the first node matching the selector belongs to.
func (c *Puppet) Submit(sel string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.Submit(sel))
}

// SetUploadFiles sets the files to upload (ie, for a input[type="file"] node) for the first node matching the selector.
func (c *Puppet) SetUploadFiles(sel string, files []string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.SetUploadFiles(sel, files))
}

// Reset is an action that resets the form of the first node matching the selector belongs to.
func (c *Puppet) Reset(sel string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.Reset(sel))
}

// ScrollIntoView scrolls the window to the first node matching the selector.
func (c *Puppet) ScrollIntoView(sel string) (err error) {
	return c.cdp.Run(c.ctx,
		chromedp.ScrollIntoView(sel))
}

// SetHeaders specifies whether to always send extra HTTP headers with the requests from this page.
func (c *Puppet) SetHeaders(headers map[string]interface{}) (err error) {
	return c.cdp.Run(c.ctx,
		network.SetExtraHTTPHeaders(network.Headers(headers)))
}

// SetCookies sets given cookies.
func (c *Puppet) SetCookies(cookies []*http.Cookie) (err error) {
	cookieParams := []*network.CookieParam{}
	for _, cookie := range cookies {
		expr := cdp.TimeSinceEpoch(cookie.Expires)
		var cookieSameSite network.CookieSameSite
		switch cookie.SameSite {
		case http.SameSiteDefaultMode:
		case http.SameSiteLaxMode:
			cookieSameSite = network.CookieSameSiteLax
		case http.SameSiteStrictMode:
			cookieSameSite = network.CookieSameSiteStrict
		}
		cookieParams = append(cookieParams, &network.CookieParam{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Secure:   cookie.Secure,
			HTTPOnly: cookie.HttpOnly,
			SameSite: cookieSameSite,
			Expires:  &expr,
		})
	}

	err = c.cdp.Run(c.ctx,
		network.SetCookies(cookieParams))
	if err != nil {
		return err
	}
	return nil
}

// DelCookies deletes browser cookies with matching name and url or domain/path pair.
func (c *Puppet) DelCookies(name string) (err error) {
	return c.cdp.Run(c.ctx,
		network.DeleteCookies(name))
}

// ClearCookies clears browser cookies.
func (c *Puppet) ClearCookies() (err error) {
	return c.cdp.Run(c.ctx,
		network.ClearBrowserCookies())
}

// Cookies returns all browser cookies. Depending on the backend support, will return detailed cookie information in the cookies field.
func (c *Puppet) Cookies() (cookies []*http.Cookie, err error) {
	err = c.cdp.Run(c.ctx, chromedp.ActionFunc(func(ctxt context.Context, h cdp.Executor) error {
		cookieResults, err := network.GetAllCookies().
			Do(ctxt, h)
		if err != nil {
			return err
		}
		for _, cookie := range cookieResults {
			var cookieSameSite = http.SameSiteDefaultMode
			switch cookie.SameSite {
			case network.CookieSameSiteLax:
				cookieSameSite = http.SameSiteLaxMode
			case network.CookieSameSiteStrict:
				cookieSameSite = http.SameSiteStrictMode
			}
			cookies = append(cookies, &http.Cookie{
				Name:     cookie.Name,
				Value:    cookie.Value,
				Domain:   cookie.Domain,
				Path:     cookie.Path,
				Secure:   cookie.Secure,
				HttpOnly: cookie.HTTPOnly,
				SameSite: cookieSameSite,
				Expires:  time.Date(1970, 1, 1, 0, 0, int(cookie.Expires), 0, time.UTC).Local(),
			})
		}
		return nil
	}))
	return cookies, err
}

// PDF print page as PDF.
func (c *Puppet) PDF() (res []byte, err error) {
	err = c.cdp.Run(c.ctx, chromedp.ActionFunc(func(ctxt context.Context, h cdp.Executor) error {
		res, err = page.PrintToPDF().
			WithMarginTop(0.01).
			WithMarginBottom(0.01).
			WithMarginRight(0.01).
			WithMarginLeft(0.01).
			WithPreferCSSPageSize(true).
			WithPrintBackground(true).
			WithLandscape(true).
			Do(ctxt, h)
		return err
	}),
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Screenshot capture page screenshot.
func (c *Puppet) Screenshot() (res []byte, err error) {
	err = c.cdp.Run(c.ctx, chromedp.ActionFunc(func(ctx context.Context, h cdp.Executor) error {
		res, err = page.CaptureScreenshot().
			Do(ctx, h)
		return err
	}),
	)

	if err != nil {
		return nil, err
	}
	return res, nil
}

// Snapshot returns a snapshot of the page as a string. For MHTML
// format, the serialization includes iframes, shadow DOM, external resources,
// and element-inline styles.
func (c *Puppet) Snapshot() (res []byte, err error) {
	var src string
	err = c.cdp.Run(c.ctx, chromedp.ActionFunc(func(ctx context.Context, h cdp.Executor) error {
		src, err = page.CaptureSnapshot().
			Do(ctx, h)
		return err
	}),
	)
	if err != nil {
		return nil, err
	}
	return *(*[]byte)(unsafe.Pointer(&src)), nil
}

// ClearCache clears browser cache.
func (c *Puppet) ClearCache() (err error) {
	return c.cdp.Run(c.ctx,
		network.ClearBrowserCache())
}

var waitComplete = chromedp.ActionFunc(func(ctx context.Context, h cdp.Executor) error {
	state := ""
	for i := 0; i != 10; i++ {
		if err := readyState(&state).Do(ctx, h); err != nil {
			return err
		}
		if state == "complete" {
			break
		}
		time.Sleep(time.Second / 10 * time.Duration(i+1))
	}
	return nil
})

func readyState(state *string) chromedp.Action {
	if state == nil {
		panic("state cannot be nil")
	}
	return chromedp.EvaluateAsDevTools(`document.readyState`, state)
}
