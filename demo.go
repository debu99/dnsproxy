// Package dnsproxy implements a plugin
package dnsproxy

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/coredns/coredns/request"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("dnsproxy")

// Dnsproxy is a plugin in CoreDNS
type Dnsproxy struct{}

// ServeDNS implements the plugin.Handler interface.
func (p Dnsproxy) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	fmt.Println("example0") 

	// Debug log that we've have seen the query. This will only be shown when the debug plugin is loaded.
	log.Debug("Received response")

	// Wrap.
	pw := NewResponsePrinter(w)

	state := request.Request{W: w, Req: r}
	qname := state.Name()

	reply := "8.8.8.8"
	if strings.HasPrefix(state.IP(), "172.") || strings.HasPrefix(state.IP(), "127.") || strings.HasPrefix(state.IP(), "192.") {
		reply = "1.1.1.1"
	}
	fmt.Printf("Received query %s from %s, expected to reply %s\n", qname, state.IP(), reply)

	answers := []dns.RR{}

	if state.QType() != dns.TypeA {
		return dns.RcodeNameError, nil
	}

	rr := new(dns.A)
	rr.Hdr = dns.RR_Header{Name: qname, Rrtype: dns.TypeA, Class: dns.ClassINET}
	rr.A = net.ParseIP(reply).To4()

	answers = append(answers, rr)

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = answers

	w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (p Dnsproxy) Name() string { return "dnsproxy" }

// ResponsePrinter wrap a dns.ResponseWriter and will write example to standard output when WriteMsg is called.
type ResponsePrinter struct {
	dns.ResponseWriter
}

// NewResponsePrinter returns ResponseWriter.
func NewResponsePrinter(w dns.ResponseWriter) *ResponsePrinter {
	return &ResponsePrinter{ResponseWriter: w}
}

// WriteMsg calls the underlying ResponseWriter's WriteMsg method and prints "example" to standard output.
func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	log.Info("example")
	return r.ResponseWriter.WriteMsg(res)
}

