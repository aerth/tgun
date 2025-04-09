// Copyright 2017 The tgun Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tgun

import (
	"encoding/json"
	"fmt"
	"net"
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

const skiptest = "SKIPPED. Set TESTALL in your environment to run this test."

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
		if i < 10 {
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
		if r.UserAgent() != DefaultUserAgent {
			fmt.Println("It should be:", DefaultUserAgent)
			fmt.Printf("Your User Agent is: %q\n", r.UserAgent())
			fmt.Println(r.URL.String())
			t.FailNow()
		}
		i++
		if i < 10 {
			http.Redirect(w, r, fmt.Sprintf("/?test-forwarding-%d", i), http.StatusFound)
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))

	_, err := dialer.GetBytes(ts.URL)
	if err != nil {
		fmt.Println(err)
	}
}

func TestHeaders(t *testing.T) {
	dialer := &Client{
		Proxy:     "", // direct
		UserAgent: "", // defaultUserAgent
		Headers: map[string]string{
			"MyHeader1": "123", // note capitalization of key funk
			"MyHeader2": "456",
		},
	}

	b, err := dialer.GetBytes("https://httpbin.org/headers")
	if err != nil {
		fmt.Println(err)
		t.FailNow()
		return
	}

	reply := struct {
		Headers map[string]string `json:"headers"`
	}{}
	err = json.Unmarshal(b, &reply)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(string(b))
		fmt.Println(err)
		t.FailNow()
		return
	}

	if reply.Headers["Myheader1"] != "123" { // Note capitalization difference in reply funk
		fmt.Printf("Expected MyHeader1 to be 123, got %q\n", reply.Headers["MyHeader1"])
		t.Fail()
	}
	if reply.Headers["Myheader2"] != "456" {
		fmt.Printf("Expected MyHeader2 to be 456, got %q\n", reply.Headers["MyHeader2"])
		t.Fail()
	}
}

func TestTor(t *testing.T) {
	if os.Getenv("TESTALL") == "" {
		fmt.Println(skiptest)
		return
	}
	dialer := &Client{
		Proxy:     "tor",
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
	torline := "Congratulations. This browser is configured to use Tor."
	if strings.Contains(string(b), torline) {
		fmt.Println("Success:", torline)
		return
	}

	fmt.Println(string(b))
	t.Fail()
}

func TestSOCKS(t *testing.T) {
	if os.Getenv("TESTALL") == "" {
		fmt.Println(skiptest)
		return
	}
	dialer := &Client{
		Proxy:     "socks5h://127.0.0.1:1080",
		UserAgent: "Testing/1.2",
	}

	b, err := dialer.GetBytes("http://icanhazip.com")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Test Failed. Check your connection and make sure SOCKS5 proxy is listening on 1080.")
		t.FailNow()
	}

	fmt.Println("Success on 1080:", string(b))

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

	dialer.Proxy = "example.com" // wrong format, needs socks5h://socks.com
	_, err = dialer.GetBytes(ts.URL)
	if err == nil {
		t.FailNow()
	}

	dialer.Proxy = "example.org:1080" // wrong format, needs socks5h://socks.com
	_, err = dialer.GetBytes(ts.URL)
	if err == nil {
		t.FailNow()
	}

}

// TestDirect actually direct connects (no proxy)
func TestDirect(t *testing.T) {
	dialer := &Client{
		UserAgent: "Testing/1.2",
	}
	handler := func(w http.ResponseWriter, r *http.Request) {}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	_, err := dialer.GetBytes(ts.URL)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
}

// TestSimpleAuth sends auth information with request
func TestSimpleAuth(t *testing.T) {
	dialer := &Client{
		Proxy:        "",
		UserAgent:    "Testing/1.2",
		AuthUser:     "user1",
		AuthPassword: "password321",
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			t.Log("Expected to use user1:password321 for authenticated request, got \"not okay\"")
			t.Fail()
			return

		}
		if user != "user1" || pass != "password321" {
			t.Log("Expected to use user1:password321 for authenticated request, got:", user, pass)
			t.Fail()
			return
		}

	}
	ts := httptest.NewServer(http.HandlerFunc(handler))

	_, err := dialer.GetBytes(ts.URL)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
}

// TestBadRequest returns an error before connection
func TestBadRequest(t *testing.T) {
	badurl := "www.example.org" // should have a protocol scheme (such as http)
	dialer := &Client{}
	_, err := dialer.Get(badurl)
	if err == nil {
		t.Log("Expected an error, got none.")
		t.FailNow()
	}
	t.Log(t.Name(), "Pass:", err)

}

type githubRepoT struct {
	ID              int    `json:"id"`
	NodeID          string `json:"node_id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	URL             string `json:"url"`
	Size            int    `json:"size"`
	StargazersCount int    `json:"stargazers_count"`
	WatchersCount   int    `json:"watchers_count"`
	//	SubscribersCount int    `json:"subscribers_count"`
}

// curl -H "Accept: application/vnd.github.v3+json" \
//   https://api.github.com/repos/octocat/hello-world
func TestJSON(t *testing.T) {
	client := &Client{}
	endpoint := "api.github.com"
	u := "https://" + client.Join(endpoint, "repos", "aerth", "tgun")
	var githubRepo githubRepoT
	if err := client.Unmarshal(u, &githubRepo); err != nil {
		t.Errorf("%v", err)
		return
	}
	fmt.Printf("%s (%d ★ %d 👓)\n",
		githubRepo.Name,
		githubRepo.StargazersCount,
		githubRepo.WatchersCount)
}

func TestJoin(t *testing.T) {
	type testcase []string
	// "https://" + Join("api.example.com/v1", "users")
	// "https://" + Join("api.example.com", "v1", "users")
	// Join("http://", "example.com", "index.php")
	// Join("http://api.example.com/v1", "users")
	// Join("http://api.example.com", "v1", "users")

	for _, tc := range []testcase{
		testcase{"api.example.com/v1", "users", "api.example.com/v1/users"},
		testcase{"api.example.com", "v1", "users", "api.example.com/v1/users"},
		testcase{"https://", "api.example.com/v1", "users", "https://api.example.com/v1/users"},
		testcase{"https://api.example.com/v1", "users", "https://api.example.com/v1/users"},
		testcase{"https://api.example.com", "v1", "users", "https://api.example.com/v1/users"},
	} {
		//
		i := len(tc) - 1
		got := Join(tc[:i]...)
		want := string(tc[i])
		if got != want {
			t.Errorf("wanted %q, got %q", want, got)
		}
		println(got)
	}
}

func TestDialer(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	ch := make(chan []byte)
	go func(ch chan []byte) {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		buf := make([]byte, 4)
		n, err := conn.Read(buf)
		if err != nil {
			panic(err)
		}
		ch <- buf[:n]
	}(ch)
	c := &Client{Proxy: ""}
	conn, err := c.DialTCP("127.0.0.1:8080")
	if err != nil {
		t.Errorf("err: %v", err)
		return
	}
	_, err = conn.Write([]byte("hi\n"))
	if err != nil {
		t.Errorf("err: %v", err)
		return
	}
	got := <-ch
	t.Logf("got: %s (%02x)", string(got), got)
	if string(got) != "hi\n" {
		t.FailNow()
	}
}
