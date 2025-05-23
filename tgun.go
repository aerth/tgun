// Copyright 2017 The tgun Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tgun provides a TCP/http(s) client with common options
package tgun

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

const version = "0.2.0"

// Version of tgun library
func Version() string {
	return version
}

// DefaultTimeout is used if c.Timeout is not set
var DefaultTimeout = time.Second * 30

// DefaultProxy is used when c.Proxy is "1080" or "socks" or "proxy" or "true" or "1"
var DefaultProxy = "socks5h://127.0.0.1:1080"

// DefaultUserAgent is used when c.UserAgent is empty
var DefaultUserAgent = fmt.Sprintf("tgun/%s", version)

// DefaultTor proxy is used when c.Proxy is set to "tor"
var DefaultTor = func() string {
	if runtime.GOOS == "windows" {
		return "socks5h://127.0.0.1:9150"
	}

	return "socks5h://127.0.0.1:9050"
}()

// Client holds connection options
type Client struct {
	DirectConnect bool          // Set to true to bypass proxy
	Proxy         string        // In the format: socks5h://localhost:1080
	UserAgent     string        // In the format: "MyThing/0.1" or "MyThing/0.1 (http://example.org)"
	Timeout       time.Duration // If unset, DefaultTimeout is used.
	AuthUser      string
	AuthPassword  string
	Headers       map[string]string
	httpClient    *http.Client
	dialer        proxy.Dialer
}

// HTTPClient returns a http.Client with proxy (but no headers, auth, user-agent)
func (c Client) HTTPClient() (*http.Client, error) {
	err := c.refresh()
	return c.httpClient, err
}

func Join(s ...string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) == 1 {
		return s[0]
	}

	// "https://" + Join("api.example.com/v1", "users")
	// "https://" + Join("api.example.com", "v1", "users")
	if !strings.Contains(s[0], "://") {
		return path.Join(s...)
	}

	// Join("http://", "example.com", "index.php")
	// Join("http://api.example.com/v1", "users")
	// Join("http://api.example.com", "v1", "users")

	if strings.HasSuffix(s[0], "://") {
		s[1] = s[0] + s[1]
		s = s[1:]
	}

	u, err := url.Parse(s[0])
	if err != nil {
		return ""
	}
	u.Path = path.Join(append([]string{u.Path}, s[1:]...)...)
	return u.String()
}

func (c *Client) Join(s ...string) string {
	return Join(s...)
}

// DialTCP an address using c.Proxy if set
func (c *Client) DialTCP(addr string) (net.Conn, error) {
	return c.Dial("tcp", addr)
}

// Dial an address using c.Proxy if set
func (c *Client) Dial(network string, addr string) (net.Conn, error) {
	// Refresh http client, proxy
	if err := c.refresh(); err != nil {
		return nil, err
	}
	return c.dialer.Dial(network, addr)
}

// Do returns an http response.
// The request's config is *fortified* with http.Client, proxy, headers, authentication, and user agent.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Refresh http client, proxy
	if err := c.refresh(); err != nil {
		return nil, err
	}

	// Create request headers using tgun.Client config
	if c.Headers != nil {
		for header, value := range c.Headers {
			req.Header.Set(header, value)
		}
	}

	// Set Basic Auth
	if c.AuthUser != "" && c.AuthPassword != "" {
		req.SetBasicAuth(c.AuthUser, c.AuthPassword)
	}

	// Set User Agent, over previous headers
	req.Header.Set("User-Agent", c.UserAgent)

	// Do with new http client request
	return c.httpClient.Do(req)
}

// Unmarshal JSON (GET, no body)
func (c *Client) Unmarshal(url string, ptr interface{}) error {
	return c.unmarshal(url, http.MethodGet, ptr, nil)
}

// Unmarshal JSON (POST, with body param)
func (c *Client) UnmarshalPost(url string, ptr interface{}, body io.Reader) error {
	return c.unmarshal(url, http.MethodPost, ptr, body)
}

func (c *Client) unmarshal(url string, method string, ptr interface{}, body io.Reader) error {
	if err := c.refresh(); err != nil {
		return err
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return err
	}
	err = json.Unmarshal(b, ptr)
	if err != nil {
		resp.Body.Close()
		return fmt.Errorf("%d: %s", resp.StatusCode, string(b))
	}
	resp.Body.Close()
	return nil
}

// Get connects returns an http response
func (c *Client) Get(url string) (*http.Response, error) {
	if err := c.refresh(); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

// GetBytes connects and returns an http response body in the form of bytes
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

// getDialer is called by proxify to return a proxy.Dialer
func getDialer(proxyurl string) (proxy.Dialer, error) {
	switch proxyurl {
	case "":
		return proxy.Direct, nil
	case "tor":
		proxyurl = DefaultTor
	case "socks", "1080", "proxy", "true", "1":
		proxyurl = DefaultProxy
	}

	u, err := url.Parse(proxyurl)
	if err != nil {
		return nil, err
	}

	return proxy.FromURL(u, proxy.Direct)
}

// proxify is called by refresh to return a *http.Transport
func (c *Client) proxify() (*http.Transport, error) {
	t := &http.Transport{}
	proxypath := c.Proxy
	if c.DirectConnect {
		proxypath = ""
	}
	dialer, err := getDialer(proxypath)
	if err != nil {
		return nil, fmt.Errorf("Dialer Error: %s", err.Error())
	}
	c.dialer = dialer
	t.Dial = c.dialer.Dial
	return t, nil
}

// refresh gets called every time we use any method of Client
// its responsibility is:
//
//		to sanity check the tgun.Client config
//	 to apply c.Proxy, and fix redirect useragent/header leak.
func (c *Client) refresh() error {
	// default user agent
	if c.UserAgent == "" {
		c.UserAgent = DefaultUserAgent
	}

	// default timeout
	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}

	// create transport
	transport, err := c.proxify()
	if err != nil {
		return err
	}

	// forge http client
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   c.Timeout,
		Jar:       nil,
	}

	// create redirect policy to set UA even during redirects
	httpClient.CheckRedirect = func(req *http.Request, reqs []*http.Request) error {
		for h, v := range c.Headers {
			req.Header.Set(h, v)
		}
		req.Header.Set("User-Agent", c.UserAgent)
		return nil
	}

	c.httpClient = httpClient

	return nil
}
