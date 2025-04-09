package main

import (
	"io"
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

//export version
func version() *C.char {
	println("deprecated: use tgunversion()")
	return C.CString(tgun.Version())
}

//export tgunversion
func tgunversion() *C.char {
	return C.CString(tgun.Version())
}

// sets proxy string for future requests (example: "socks5h://127.0.0.1:1080" or "tor" or "socks" or "")
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

func toMethod(s string) string {
	s = strings.ToUpper(s)
	switch s {
	case "", http.MethodGet:
		return http.MethodGet
	default:
		return s
	}
}

// tgun_do copies response body directly to file instead of memory
//
//export tgun_do
func tgun_do(method *C.char, url *C.char, headers *C.char, output_filename *C.char) int {
	//fmt.Printf("tgun_do: url=%q m=%q h=%q o=%q\n", C.GoString(url), C.GoString(method), C.GoString(headers), C.GoString(output_filename))
	headermap := parseHeaders(headers)
	t := tgun.Client{
		Headers:   headermap,
		Proxy:     state.proxy,
		UserAgent: state.useragent,
	}
	req, err := http.NewRequest(toMethod(C.GoString(method)), C.GoString(url), nil)
	if err != nil {
		state.err = err
		return 1
	}

	resp, err := t.Do(req)
	if err != nil {
		state.err = err
		return 1
	}

	defer resp.Body.Close()
	var outputFile *os.File
	var outputFilename string
	if output_filename != nil {
		outputFilename = C.GoString(output_filename)
	}
	switch outputFilename {
	case "", "1", "stdout":
		outputFile = os.Stdout
	case "2", "stderr":
		outputFile = os.Stderr
	default:
		outputFile, err = os.OpenFile(outputFilename, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			state.err = err
			return 1
		}
		defer outputFile.Close()
	}
	_, err = io.Copy(outputFile, resp.Body)
	if err != nil {
		state.err = err
		return 1
	}

	return 0
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
	b, err := io.ReadAll(resp.Body)
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
