// package tgun provides a TCP/http(s) client with common options
package tgun

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

const (
	defaultUserAgent = "aerth_tgun/0.1"
	defaultProxy     = "socks5://127.0.0.1:1080"
	defaultTor       = "socks5://127.0.0.1:9050"
)

type Client struct {
	DirectConnect bool   // Set to true to bypass proxy
	Proxy         string // In the format: socks5://localhost:1080
	UserAgent     string // In the format: "MyThing/0.1" or "MyThing/0.1 (http://example.org)"
	Timeout       time.Duration
	Headers       map[string]string
	httpClient    *http.Client
	dialer        proxy.Dialer
	mu            sync.RWMutex
}

func (c *Client) Get(url string) (*http.Response, error) {
	if err := c.refresh(); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if c.Headers != nil {
		for k, v := range c.Headers {
			req.Header.Set(k, v)
		}
	}
	req.Header.Set("User-Agent", c.UserAgent) // should go first?
	return c.httpClient.Do(req)
}

func (c *Client) GetBytes(url string) ([]byte, error) {
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	return b, err
}

func getDialer(proxyurl string) (proxy.Dialer, error) {
	// check for keywords
	if proxyurl == "" {
		return proxy.Direct, nil
	}
	if proxyurl == "tor" {
		proxyurl = defaultTor
	}

	u, err := url.Parse(proxyurl)
	if err != nil {
		return nil, err
	}

	px, err := proxy.FromURL(u, proxy.Direct)
	if err != nil {
		return nil, err
	}

	return px, nil
}

func (c *Client) getTransport() (*http.Transport, error) {
	t := &http.Transport{}
	dialer, err := getDialer(c.Proxy)
	if err != nil {
		return nil, fmt.Errorf("Dialer Error: %s", err.Error())
	}
	c.dialer = dialer
	t.Dial = c.dialer.Dial

	return t, nil
}

func (c *Client) refresh() error {

	if c.UserAgent == "" {
		c.UserAgent = defaultUserAgent
	}

	transport, err := c.getTransport()
	if err != nil {
		return err
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   c.Timeout,
		Jar:       nil,
	}

	redirectPolicyFunc := func(req *http.Request, reqs []*http.Request) error {
		req.Header.Set("User-Agent", c.UserAgent)
		return nil
	}

	// set UA
	c.httpClient = httpClient

	// set UA even during redirects
	c.httpClient.CheckRedirect = redirectPolicyFunc
	return nil
}
