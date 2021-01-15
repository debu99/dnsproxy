// Package dnsproxy implements a plugin
package dnsproxy

import (
	"context"
	//"fmt"
	"net"
	//"strings"

	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/go-redis/redis/v8"
	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("dnsproxy")

// Dnsproxy is a plugin in CoreDNS
type Dnsproxy struct{}

// ServeDNS implements the plugin.Handler interface.
func (p Dnsproxy) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// Debug log that we've have seen the query. This will only be shown when the debug plugin is loaded.
	log.Debug("Received from plugin...")

	opt, err := redis.ParseURL("redis://localhost:6379/0")
	if err != nil {
		log.Fatal("redis url error", err)
	}

	rdb := redis.NewClient(opt)
	err = rdb.Ping(ctx).Err()
	if err != nil {
		log.Fatal("redis connect failed", err)
	}
	defer rdb.Close()

	state := request.Request{W: w, Req: r}
	qname := state.Name()

	reply := "8.8.8.8"
	ret := rdb.SIsMember(ctx, "ipsets", state.IP())
	if ret.Err() != nil {
		log.Fatal(ret.Err())
	}
	if ret.Val() {
		reply = "1.1.1.1"
	}

	log.Info("Received query %s from %s, expected to reply %s\n", qname, state.IP(), reply)

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
