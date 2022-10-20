package mserver

import (
	"fmt"
	"net"
	"os"
	"time"

	"context"

	"github.com/go-redis/redis/v8"
	"github.com/miekg/dns"
	"github.com/patrickmn/go-cache"
)

var serverurl string
var c *cache.Cache
var question *cache.Cache
var redisdb *redis.Client
var ctx context.Context

func Listen(port int, serveraddr string) {
	serverurl = serveraddr
	serveMux := dns.NewServeMux()
	serveMux.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {
		handleRequest(w, req)
	})
	c = cache.New(5*time.Minute, 10*time.Minute)
	question = cache.New(2*time.Second, 2*time.Minute)

	// Now for redis
	ctx = context.Background()
	redisdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

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

				//fmt.Printf("From cache: %+v\n", r.Question)
				m.Answer = append(m.Answer, res.(dns.Msg).Answer...)
				m.Ns = append(m.Ns, res.(dns.Msg).Ns...)
				m.Extra = append(m.Extra, res.(dns.Msg).Extra...)
				w.WriteMsg(m)
				return
			}
			d := net.Dialer{Timeout: 500 * time.Millisecond}
			conn, err := d.Dial("udp", serverurl)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				return
			}
			// Check if we already asked and waiting for an Answer
			res, found = question.Get(r.Question[0].String())
			if found {
				// Then don't ask again, fail happily
				// TODO: I don't know how bad is this idea :)
				w.WriteMsg(m)
				return
			}
			// Now we record that we asked this question
			question.Set(r.Question[0].String(), true, cache.DefaultExpiration)
			//fmt.Printf("%+v\n", r.Question)
			defer conn.Close()
			dnsConn := &dns.Conn{Conn: conn}
			if err = dnsConn.WriteMsg(r); err != nil {
				w.WriteMsg(m)
				fmt.Fprintf(os.Stderr, "Error while talking to the server %s\n", err)
				return
			}
			resp, err := dnsConn.ReadMsg()
			if err == nil {
				// First delete from the question
				question.Delete(r.Question[0].String())
				m.Answer = append(m.Answer, resp.Answer...)
				m.Ns = append(m.Ns, resp.Ns...)
				m.Extra = append(m.Extra, resp.Extra...)
				if len(resp.Answer) > 0 {
					//fmt.Printf("TTL: %+v\n", resp.Answer[0].Header().Ttl)
					c.Set(r.Question[0].String(), *resp, cache.DefaultExpiration)
					go pushToRedis(r, resp)
				}
			}
		}
	}

	w.WriteMsg(m)
}

func pushToRedis(r *dns.Msg, answer *dns.Msg) {

	// For each IP, record the DNS name
	for _, ans := range answer.Answer {
		if ip, ok := ans.(*dns.A); ok {
			rname := fmt.Sprintf("ip:%s", ip.A.String())
			redisdb.SAdd(ctx, rname, r.Question[0].Name)
		}
		if ip, ok := ans.(*dns.AAAA); ok {
			rname := fmt.Sprintf("ip:%s", ip.AAAA.String())
			redisdb.SAdd(ctx, rname, r.Question[0].Name)
		}
	}

}
