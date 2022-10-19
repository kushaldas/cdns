package mserver

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/miekg/dns"
	"github.com/patrickmn/go-cache"
)

var serverurl string
var c *cache.Cache

func Listen(port int, serveraddr string) {
	serverurl = serveraddr
	serveMux := dns.NewServeMux()
	serveMux.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {
		handleRequest(w, req)
	})
	c = cache.New(5*time.Minute, 10*time.Minute)

	server := &dns.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port), Net: "udp", Handler: serveMux}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while starting the server: %s\n", err)
		os.Exit(127)
	}

}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	if r.MsgHdr.Opcode == dns.OpcodeQuery {
		if len(r.Question) > 0 {

			// First check the cache
			res, found := c.Get(r.Question[0].String())
			if found {

				fmt.Printf("From cache: %+v\n", r.Question)
				m.Answer = append(m.Answer, res.(dns.Msg).Answer...)
				m.Ns = append(m.Ns, res.(dns.Msg).Ns...)
				m.Extra = append(m.Extra, res.(dns.Msg).Extra...)
				w.WriteMsg(m)
				return
			}
			d := net.Dialer{Timeout: 3 * time.Second}
			conn, err := d.Dial("udp", serverurl)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				return
			}
			fmt.Printf("%+v\n", r.Question)
			defer conn.Close()
			dnsConn := &dns.Conn{Conn: conn}
			if err = dnsConn.WriteMsg(r); err != nil {
				w.WriteMsg(m)
				fmt.Fprintf(os.Stderr, "Error while talking to the server %s\n", err)
				return
			}
			resp, err := dnsConn.ReadMsg()
			if err == nil {
				m.Answer = append(m.Answer, resp.Answer...)
				m.Ns = append(m.Ns, resp.Ns...)
				m.Extra = append(m.Extra, resp.Extra...)
				if len(resp.Answer) > 0 {
					//fmt.Printf("TTL: %+v\n", resp.Answer[0].Header().Ttl)
					c.Set(r.Question[0].String(), *resp, cache.DefaultExpiration)
				}
			}
		}
	}

	w.WriteMsg(m)
}
