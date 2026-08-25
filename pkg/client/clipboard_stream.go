package client

import (
	"context"
	"encoding/json"
	"github.com/peardesk/peardesk/pkg/clipboard"
	"github.com/peardesk/peardesk/pkg/protocol"
	"github.com/peardesk/peardesk/pkg/tor"
	"time"
)

func (c *Connection) clipLoop(ctx context.Context) {
	m := clipboard.New(500 * time.Millisecond)
	c.cbMon = m
	m.Start(func(t string) {
		b, _ := json.Marshal(protocol.ClipboardMsg{Type: protocol.TypeClipboard, Text: t})
		_ = c.send(tor.FrameClipboard, b)
	})
	<-ctx.Done()
}
func (c *Connection) SendClipboard(t string) {
	b, _ := json.Marshal(protocol.ClipboardMsg{Type: protocol.TypeClipboard, Text: t})
	_ = c.send(tor.FrameClipboard, b)
}
