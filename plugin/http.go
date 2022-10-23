package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"unsafe"

	"github.com/aerth/tgun"
)

/*
#include <stdlib.h>
*/
import "C"

// global in your application for now
var state = struct {
	proxy     string
	useragent string
	err       error
	errs      *C.char
}{
	proxy:     os.Getenv("PROXY"),
	useragent: os.Getenv("USER_AGENT"),
}

var Version = "0.0.1"

//export version
func version() *C.char {
	return C.CString(Version)
}

// sets proxy string for future requests (example: "socks5://127.0.0.1:1080" or "tor")
//
//export easy_proxy
func easy_proxy(proxystring *C.char) {
	state.proxy = C.GoString(proxystring)
}

// sets user agent for future requests
//
//export easy_ua
func easy_ua(useragent *C.char) {
	state.useragent = C.GoString(useragent)
}

// displays latest error *and clears it*
// DONT free it, it gets freed next call, so just call it again as a void fn.
//
// not safe in general.
//
//export tgunerr
func tgunerr() *C.char {
	if state.errs != nil { // definitely old err
		C.free(unsafe.Pointer(state.errs))
	}
	if state.err == nil {
		return nil
	}
	// get current err string and clear the err
	es := state.err.Error()
	state.err = nil

	esc := C.CString(es)
	state.errs = esc
	return esc
}

// (remember to free())
//
//export get_url
func get_url(url *C.char) *C.char {
	return get_url_headers(url, nil)
}

// (remember to free())
//
//export get_url_headers
func get_url_headers(url *C.char, headers *C.char) *C.char {
	headermap := parseHeaders(headers)
	t := tgun.Client{
		Headers:   headermap,
		Proxy:     state.proxy,
		UserAgent: state.useragent,
	}
	b, err := t.GetBytes(C.GoString(url))
	if err != nil {
		state.err = err
		return nil
	}
	return C.CString(string(b)) // todo: handle files etc
}

// special header format: content-type=application/json;accept=*;
// or multiple: accept=foo=bar
// (remember to free())
//
//export post_url
func post_url(url *C.char, bodyString *C.char, headers *C.char) *C.char {
	u := C.GoString(url)
	headermap := parseHeaders(headers)
	t := tgun.Client{
		Headers:   headermap,
		Proxy:     state.proxy,
		UserAgent: state.useragent,
	}
	req, err := http.NewRequest(http.MethodPost, u, strings.NewReader(C.GoString(bodyString)))
	if err != nil {
		state.err = err
		return nil
	}
	resp, err := t.Do(req)
	if err != nil {
		state.err = err
		return nil
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		state.err = err
		return nil
	}
	return C.CString(string(b))
}

// special header format see above function exported comment
func parseHeaders(headers *C.char) map[string]string {
	hh := strings.Split(C.GoString(headers), ";") // only for now?
	if hh[0] == "" {
		return nil
	}
	headermap := map[string]string{}
	for _, h := range hh {
		spl := strings.Split(h, "=")
		//log.Println("header:", h, spl)
		headermap[spl[0]] = strings.Join(spl[1:], ";") // only for now?
	}
	return headermap
}

func main() {
	panic("this is a plugin")
}
