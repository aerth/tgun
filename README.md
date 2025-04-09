# tgun

### a http and tcp client with common options

  * Use **Proxy** (http, socks4, socks5, tor)
  * Use custom **UserAgent** (even during redirects)
  * Set **headers**
  * Use **simple authentication**
  * Custom timeout

```
// set headers if necessary
headers := map[string]string{
  "API_KEY": "12345"
  "API_SECRET": "12345"
}

// set user agent and proxy in the initialization
dialer := tgun.Client{
  Proxy:     "socks5h://localhost:1080",
  UserAgent: "MyCrawler/0.1 (https://github.com/user/repo)",
  Headers:   headers,
}

// get bytes
b, err := dialer.GetBytes("https://example.org")

```

See [tgun_test.go](tgun_test.go) for more examples.

### c usage

harness tgun in your c application!

first `make` in plugin directory, creating `tgun.a tgun.so tgun.h` and an example `tgun` curl-like application.

```
#include <tgun.h>

int main(){
    // set user-agent
    easy_ua("libtgun/1.0");
    // set proxy url, or alias 'tor' (9050 or 9150 depending on platform) or 'socks' (127.0.0.1:1080)
    easy_proxy("tor");
    char* b = get_url("http://example.org");

    // if any errors, NULL is returned and an error is waiting
    if (!b) {
      fprintf(stderr, "error: %s\n", tgunerr());    
    } else {
      // normal string, do something with it, then free().
      printf("%s", b);
      free(b);
    }
}

```

see [plugin](plugin) directory for c usage example
