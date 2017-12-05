# tgun

### a tcp client with common options

  * Use Proxy (http, socks4, socks5, tor)
  * Use Custom UserAgent (even during redirects)
  * Set headers

```
// set headers if necessary
headers := map[string]string{
  "API_KEY": "12345"
  "API_SECRET": "12345"
}

// set user agent and proxy in the initialization
dialer := tgun.Client{
  Proxy:     "socks5://localhost:1080",
  UserAgent: "CBaser/0.1 (https://github.com/aerth/cbaser)",
  Headers:   headers,
}

// get bytes
b, err := dialer.GetBytes("https://example.org")

```
