package main

// typedef const char constchar;
import "C"
import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"runtime"
	"unsafe"

	"github.com/aerth/tgun"
)

var conns = map[int]net.Conn{}
var connstruct = struct {
	s1  net.Conn
	s2  net.Conn
	s3  net.Conn
	s4  net.Conn
	s5  net.Conn
	s6  net.Conn
	s7  net.Conn
	s8  net.Conn
	s9  net.Conn
	s10 net.Conn
}{}

//export tgun_write
func tgun_write(fd int, buf *C.char) C.int {
	conn := getconn(fd)
	if conn == nil {
		state.err = fmt.Errorf("not found")
		return -1
	}
	n, err := conn.Write([]byte(C.GoString(buf))) // todo no copy
	if err == nil {
		return C.int(n)
	}
	state.err = err
	return -1
}

//export tgun_read
func tgun_read(fd int, buf *C.char, maxwrite C.int) (nread C.int) {
	conn := getconn(fd)
	if conn == nil {
		state.err = fmt.Errorf("not found")
		return -1
	}
	cbuf := (*[1 << 28]byte)(unsafe.Pointer(buf))[:maxwrite:maxwrite]
	//log.Printf("reading cbuf %02x", cbuf)
	n, err := conn.Read(cbuf)
	if err == nil {
		return C.int(n)
	}
	state.err = err
	if n > 0 {
		return C.int(n)
	}
	return -1
}

//export tgun_disconnect
func tgun_disconnect(fd int) {
	conn := getconn(fd)
	if conn != nil {
		conn.Close()
	}
}

func getconn(n int) net.Conn {
	switch n {
	case 0:
		return nil
	case 1:
		return connstruct.s1
	case 2:
		return connstruct.s1
	case 3:
		return connstruct.s1
	case 4:
		return connstruct.s1
	case 5:
		return connstruct.s5
	default:
		return conns[n]
	}
}

//export tgun_connect
func tgun_connect(addr *C.constchar, port C.int, dotls bool, unsafetls bool) (fd C.int) {
	tg := &tgun.Client{
		Proxy:   os.Getenv("PROXY"),
		Timeout: 0,
	}
	var conn net.Conn
	var err error
	switch {
	case dotls && unsafetls:
		conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", C.GoString(addr), int(port)), &tls.Config{
			InsecureSkipVerify: true,
		})
	case dotls:
		conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", C.GoString(addr), int(port)), &tls.Config{
			InsecureSkipVerify: false,
		})

	default:
		conn, err = tg.DialTCP(fmt.Sprintf("%s:%d", C.GoString(addr), int(port)))
	}
	if err != nil {

		state.err = err
		return -1
	}
	println("connected")
	l := len(conns) + 1
	switch l {
	case 0:
		return -1
	case 1:
		connstruct.s1 = conn
	case 2:
		connstruct.s2 = conn
	case 3:
		connstruct.s3 = conn
	case 4:
		connstruct.s4 = conn
	case 5:
		connstruct.s5 = conn
	case 6:
		connstruct.s6 = conn
	case 7:
		connstruct.s7 = conn
	case 8:
		connstruct.s8 = conn
	case 9:
		connstruct.s9 = conn
	case 10:
		connstruct.s10 = conn
	}
	conns[l] = conn
	pn.Pin(conn)
	return C.int(l)
}

var pn runtime.Pinner
