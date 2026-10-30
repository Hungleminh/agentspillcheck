// Copyright 2026 Hung Le and contributors
// SPDX-License-Identifier: Apache-2.0

// Package proxy is the egress proxy (PLAN 8.7): v0.1 records
// proxy.connect (host, port, byte counts) and tunnels bytes. It
// refuses private/internal destinations unless they are explicitly
// allowed stage hosts (SSRF guard, PLAN 17.2). MITM is v0.2.
package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Hungleminh/agentspillcheck/internal/event"
)

// Options configures the proxy.
type Options struct {
	// AllowHosts may receive traffic despite resolving privately
	// (stage services). Keys are host:port.
	AllowHosts map[string]bool
	// Transport overrides the upstream transport (used by tests so no
	// real sockets are needed). Nil uses a direct dialing transport.
	Transport http.RoundTripper
}

// Proxy is one session's egress proxy.
type Proxy struct {
	session string
	bus     *event.Bus
	opts    Options
}

// New returns a proxy for a session.
func New(sessionID string, bus *event.Bus, opts Options) *Proxy {
	return &Proxy{session: sessionID, bus: bus, opts: opts}
}

func (p *Proxy) emitConnect(host string, bytesIn, bytesOut int64) {
	ev, err := event.New(p.session, event.SourceProxy, event.KindProxyConnect)
	if err != nil {
		return
	}
	ev.Attributes["server.address"] = host
	ev.Attributes["spill.bytes_in"] = bytesIn
	ev.Attributes["spill.bytes_out"] = bytesOut
	_ = p.bus.Publish(context.Background(), ev)
}

func (p *Proxy) allowed(r *http.Request, host string) bool {
	if p.opts.AllowHosts[host] {
		return true
	}
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}
	ips, err := net.DefaultResolver.LookupIP(context.Background(), "ip", hostname)
	if err != nil {
		if ip := net.ParseIP(hostname); ip != nil {
			ips = []net.IP{ip}
		} else {
			return false
		}
	}
	for _, ip := range ips {
		if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			return false
		}
	}
	_ = r
	return true
}

// Handler returns an http.Handler acting as an HTTP proxy.
func (p *Proxy) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodConnect {
			p.handleConnect(w, r)
			return
		}
		if r.URL == nil || r.URL.Host == "" {
			http.Error(w, "proxy requires an absolute URL", http.StatusBadRequest)
			return
		}
		host := r.URL.Host
		if !p.allowed(r, host) {
			p.emitConnect(host, 0, 0)
			http.Error(w, "forbidden destination", http.StatusForbidden)
			return
		}
		p.emitConnect(host, 0, 0)
		transport := p.opts.Transport
		if transport == nil {
			transport = &http.Transport{Proxy: nil, DialContext: (&net.Dialer{Timeout: 10 * time.Second}).DialContext}
		}
		outReq := r.Clone(r.Context())
		outReq.RequestURI = ""
		resp, err := transport.RoundTrip(outReq)
		if err != nil {
			http.Error(w, "proxy upstream error", http.StatusBadGateway)
			return
		}
		defer func() { _ = resp.Body.Close() }()
		for k, vs := range resp.Header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})
}

func (p *Proxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if !p.allowed(r, host) {
		p.emitConnect(host, 0, 0)
		http.Error(w, "forbidden destination", http.StatusForbidden)
		return
	}
	target, err := net.DialTimeout("tcp", host, 10*time.Second)
	if err != nil {
		http.Error(w, "proxy dial error", http.StatusBadGateway)
		return
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		_ = target.Close()
		http.Error(w, "hijack unsupported", http.StatusInternalServerError)
		return
	}
	client, _, err := hijacker.Hijack()
	if err != nil {
		_ = target.Close()
		return
	}
	_, _ = client.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	done := make(chan struct{}, 2)
	var in, out int64
	go func() {
		n, _ := io.Copy(target, client)
		in = n
		done <- struct{}{}
	}()
	go func() {
		n, _ := io.Copy(client, target)
		out = n
		done <- struct{}{}
	}()
	<-done
	p.emitConnect(host, in, out)
	_ = client.Close()
	_ = target.Close()
}

var _ = fmt.Sprintf
var _ = strings.Contains
