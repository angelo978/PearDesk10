package host

import (
	"encoding/json"
	"github.com/peardesk/peardesk/pkg/clipboard"
	"github.com/peardesk/peardesk/pkg/protocol"
	"github.com/peardesk/peardesk/pkg/tor"
	"time"
)

func (s *Server) clipLoop(write func(byte, []byte) error) {
	m := clipboard.New(500 * time.Millisecond)
	m.Start(func(t string) {
		b, _ := json.Marshal(protocol.ClipboardMsg{Type: protocol.TypeClipboard, Text: t})
		_ = write(tor.FrameClipboard, b)
	})
}
func applyRemoteClipboard(t string) { _ = clipboard.Write(t) }
