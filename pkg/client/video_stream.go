package client

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"github.com/peardesk/peardesk/pkg/protocol"
	"github.com/peardesk/peardesk/pkg/tor"
)

func (c *Connection) readLoop(_ context.Context) {
	defer func() {
		if c.OnClose != nil {
			c.OnClose()
		}
	}()
	for {
		t, b, e := tor.ReadFrame(c.conn)
		if e != nil {
			if c.OnError != nil {
				c.OnError(e)
			}
			return
		}
		switch t {
		case tor.FrameVideo:
			if len(b) >= 8 {
				im, e := c.dec.Decode(b[8:])
				if e == nil && im != nil && c.OnFrame != nil {
					c.OnFrame(im)
				}
			}
		case tor.FrameClipboard:
			var x protocol.ClipboardMsg
			if json.Unmarshal(b, &x) == nil && c.OnClipboard != nil {
				c.OnClipboard(x.Text)
			}
		}
	}
}

var _ = binary.BigEndian
