// Copyright 2017 The tgun Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tgun

import (
	"encoding/json"
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
		if r.UserAgent() != defaultUserAgent {
			fmt.Println("It should be:", defaultUserAgent)
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
		fmt.Println(torline)
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
			t.Log("Expected to use user1:password321 for authenticated request, not okay")
			t.Fail()
			return

		}
		if user != "user1" || pass != "password321" {
			t.Log("Expected to use user1:password321 for authenticated request")
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
