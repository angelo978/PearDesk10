package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/peardesk/peardesk/pkg/clipboard"
	"github.com/peardesk/peardesk/pkg/tor"
	"github.com/peardesk/peardesk/pkg/video"
	"image"
	"net"
	"sync"
)

type Connection struct {
	conn        net.Conn
	dec         *video.Decoder
	cbMon       *clipboard.Monitor
	cancel      context.CancelFunc
	mu          sync.Mutex
	OnFrame     func(image.Image)
	OnClipboard func(string)
	OnClose     func()
	OnError     func(error)
}

func Connect(ctx context.Context, id, password string) (*Connection, error) {
	if len(id) != 62 {
		return nil, fmt.Errorf("ID Onion completo non valido")
	}
	c, e := tor.Dial(ctx, id)
	if e != nil {
		return nil, e
	}
	b, _ := json.Marshal(struct {
		Password string `json:"password"`
	}{password})
	if e = tor.WriteFrame(c, tor.FrameAuth, b); e != nil {
		c.Close()
		return nil, e
	}
	t, r, e := tor.ReadFrame(c)
	if e != nil || t != tor.FrameAuth {
		c.Close()
		return nil, fmt.Errorf("risposta Tor non valida")
	}
	var ok struct {
		OK bool `json:"ok"`
	}
	_ = json.Unmarshal(r, &ok)
	if !ok.OK {
		c.Close()
		return nil, fmt.Errorf("password errata")
	}
	d, e := video.NewDecoder()
	if e != nil {
		c.Close()
		return nil, e
	}
	x, cancel := context.WithCancel(ctx)
	z := &Connection{conn: c, dec: d, cancel: cancel}
	go z.readLoop(x)
	go z.clipLoop(x)
	return z, nil
}
func (c *Connection) Close() {
	if c.cancel != nil {
		c.cancel()
	}
	if c.cbMon != nil {
		c.cbMon.Stop()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	if c.dec != nil {
		c.dec.Close()
	}
}
func (c *Connection) send(t byte, b []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return tor.WriteFrame(c.conn, t, b)
}
func (c *Connection) Ping() {}
