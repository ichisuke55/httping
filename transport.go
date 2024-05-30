package main

import (
	"net"
	"net/http"
	"time"
)

type CustomTransport struct {
	Transport http.RoundTripper
	dialer    *net.Dialer
	reqStart  time.Time
	reqEnd    time.Time
	connStart time.Time
	connEnd   time.Time
}

func (c *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	c.reqStart = time.Now()
	resp, err := c.Transport.RoundTrip(req)
	c.reqEnd = time.Now()
	return resp, err
}

func (c *CustomTransport) dial(network, addr string) (net.Conn, error) {
	c.connStart = time.Now()
	conn, err := c.dialer.Dial(network, addr)
	c.connEnd = time.Now()
	return conn, err
}

func (c *CustomTransport) RequestDuration() time.Duration {
	return c.Duration() - c.ConnDuration()
}

func (c *CustomTransport) Duration() time.Duration {
	return c.reqEnd.Sub(c.reqStart)
}

func (c *CustomTransport) ConnDuration() time.Duration {
	return c.connEnd.Sub(c.connStart)
}

func DisableRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}
