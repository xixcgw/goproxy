package httpproxy

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/phuslu/goproxy/rootca"
	"io"
	"net"
	"net/http"
	"time"
)

type StripRequestFilter struct {
	RequestFilter
	CA *rootca.RootCA
}

func (f *StripRequestFilter) HandleRequest(h *Handler, args *FilterArgs, rw http.ResponseWriter, req *http.Request) (*http.Response, error) {
	hijacker, ok := rw.(http.Hijacker)
	if !ok {
		return nil, errors.New("http.ResponseWriter does not implments Hijacker")
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("http.ResponseWriter Hijack failed: %s", err))
	}
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	glog.Infof("%s \"STRIP %s %s %s\" - -", req.RemoteAddr, req.Method, req.Host, req.Proto)
	cert, err := f.CA.Issue(req.Host, 3*365*24*time.Hour, 2048)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("tls.LoadX509KeyPair failed: %s", err))
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
		ClientAuth:   tls.VerifyClientCertIfGiven}
	tlsConn := tls.Server(conn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		return nil, errors.New(fmt.Sprintf("tlsConn.Handshake error: %s", err))
	}
	if pl, ok := h.Listener.(PushListener); ok {
		pl.Push(tlsConn, nil)
		return nil, nil
	}
	loConn, err := net.Dial("tcp", h.Listener.Addr().String())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("net.Dial failed: %s", err))
	}
	go io.Copy(loConn, tlsConn)
	go io.Copy(tlsConn, loConn)
	return nil, nil
}

func (f *StripRequestFilter) Filter(req *http.Request) (args *FilterArgs, err error) {
	if f.CA == nil {
		return nil, errors.New("StripRequestFilter.CA is nil")
	}
	if req.Method == "CONNECT" {
		return &FilterArgs{}, nil
	}
	return nil, nil
}
