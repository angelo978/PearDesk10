package host

import (
	"encoding/json"
	"github.com/peardesk/peardesk/pkg/protocol"
)

func (s *Server) input(b []byte) {
	var m protocol.Message
	if json.Unmarshal(b, &m) != nil {
		return
	}
	w, h := screenSize()
	if w == 0 {
		w, h = 1920, 1080
	}
	switch m.Type {
	case protocol.TypeMouseEvent:
		var e protocol.MouseEventMsg
		_ = json.Unmarshal(b, &e)
		x, y := int(e.X*float64(w)), int(e.Y*float64(h))
		switch e.Action {
		case "move":
			injectMouseMove(x, y)
		case "down":
			injectMouseClick(x, y, e.Button, true)
		case "up":
			injectMouseClick(x, y, e.Button, false)
		case "scroll":
			injectMouseScroll(x, y, e.ScrollY)
		}
	case protocol.TypeKeyEvent:
		var e protocol.KeyEventMsg
		_ = json.Unmarshal(b, &e)
		injectKeyEvent(e.Key, e.Action == "down", e.Modifiers)
	case protocol.TypeRune:
		var e protocol.RuneMsg
		_ = json.Unmarshal(b, &e)
		injectRune(e.Text)
	}
}
