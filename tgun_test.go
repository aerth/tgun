package tgun

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// To run this test suite:
// Open a SOCKS proxy on 1080, start Tor on 9050.
// To complete only offline tests: go test -v
// To complete all tests: TESTALL=1 go test -v
// The test suite first tests User Agent, then checks Tor and SOCKS connections

func TestUserAgent(t *testing.T) {
	dialer := &Client{
		Proxy:     "",
		UserAgent: "Testing/1.2",
	}
	var i = 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Your User Agent is: %q\n", r.UserAgent())
		fmt.Println(r.URL.String())
		if r.UserAgent() != "Testing/1.2" {
			fmt.Println("It should be:", dialer.UserAgent)
			t.Fail()
		}
		i++
		if i < 100 {
			http.Redirect(w, r, fmt.Sprintf("/?test-forwarding-%d", i), http.StatusFound)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))

	_, err := dialer.GetBytes(ts.URL)
	if err != nil {
		fmt.Println(err)
	}
}

func TestEmptyUserAgent(t *testing.T) {
	dialer := &Client{
		Proxy:     "",
		UserAgent: "",
	}
	var i = 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Your User Agent is: %q\n", r.UserAgent())
		fmt.Println(r.URL.String())
		if r.UserAgent() != defaultUserAgent {
			fmt.Println("It should be:", defaultUserAgent)
			t.Fail()
		}
		i++
		if i < 100 {
			http.Redirect(w, r, fmt.Sprintf("/?test-forwarding-%d", i), http.StatusFound)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))

	_, err := dialer.GetBytes(ts.URL)
	if err != nil {
		fmt.Println(err)
	}
}

func TestTor(t *testing.T) {
	if os.Getenv("TESTALL") == "" {
		fmt.Println("Skipping test")
		return
	}
	dialer := &Client{
		Proxy:     "socks5://127.0.0.1:9050",
		UserAgent: "Testing/1.2",
	}

	b, err := dialer.GetBytes("https://check.torproject.org")
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			fmt.Println(err)
			fmt.Println("Test Failed. Check your connection and make sure tor is listening on port 9050.")
			t.FailNow()
		}
	}

	if strings.Contains(string(b), "Congratulations. This browser is configured to use Tor.") {
		fmt.Println("Congratulations. This browser is configured to use Tor.")
		return
	}

	fmt.Println(string(b))
	t.Fail()
}

func TestSOCKS(t *testing.T) {
	if os.Getenv("TESTALL") == "" {
		fmt.Println("Skipping test")
		return
	}
	dialer := &Client{
		Proxy:     "socks5://127.0.0.1:1080",
		UserAgent: "Testing/1.2",
	}

	b, err := dialer.GetBytes("http://icanhazip.com")
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			fmt.Println(err)
			fmt.Println("Test Failed. Check your connection and make sure SOCKS5 proxy is listening on 1080.")
			t.FailNow()
		}
	}

	fmt.Println("SUCCESS:", string(b))

}

func TestBadProxyString(t *testing.T) {
	dialer := &Client{
		Proxy:     "socks3://127.0.0.1:1080", // what is socks3?
		UserAgent: "Testing/1.2",
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		// no handle
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))

	_, err := dialer.GetBytes(ts.URL)
	if err == nil {
		t.FailNow()
	}

	dialer.Proxy = "example.com" // wrong format, needs socks5://socks.com
	_, err = dialer.GetBytes(ts.URL)
	if err == nil {
		t.FailNow()
	}

	dialer.Proxy = "example.org:1080" // wrong format, needs socks5://socks.com
	_, err = dialer.GetBytes(ts.URL)
	if err == nil {
		t.FailNow()
	}

}

func TestDirect(t *testing.T) {
	dialer := &Client{
		Proxy:     "",
		UserAgent: "Testing/1.2",
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		// no handle
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))

	_, err := dialer.GetBytes(ts.URL)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

}
