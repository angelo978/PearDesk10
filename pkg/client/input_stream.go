package client

import (
	"encoding/json"
	"github.com/peardesk/peardesk/pkg/protocol"
	"github.com/peardesk/peardesk/pkg/tor"
)

func (c *Connection) input(v interface{}) { b, _ := json.Marshal(v); _ = c.send(tor.FrameInput, b) }
func (c *Connection) SendMouseMove(x, y float64) {
	c.input(protocol.MouseEventMsg{Type: protocol.TypeMouseEvent, X: x, Y: y, Action: "move"})
}
func (c *Connection) SendMouseDown(x, y float64, b string) {
	c.input(protocol.MouseEventMsg{Type: protocol.TypeMouseEvent, X: x, Y: y, Button: b, Action: "down"})
}
func (c *Connection) SendMouseUp(x, y float64, b string) {
	c.input(protocol.MouseEventMsg{Type: protocol.TypeMouseEvent, X: x, Y: y, Button: b, Action: "up"})
}
func (c *Connection) SendScroll(x, y, d float64) {
	c.input(protocol.MouseEventMsg{Type: protocol.TypeMouseEvent, X: x, Y: y, Action: "scroll", ScrollY: d})
}
func (c *Connection) SendKeyDown(k string, m []string) {
	c.input(protocol.KeyEventMsg{Type: protocol.TypeKeyEvent, Key: k, Action: "down", Modifiers: m})
}
func (c *Connection) SendKeyUp(k string, m []string) {
	c.input(protocol.KeyEventMsg{Type: protocol.TypeKeyEvent, Key: k, Action: "up", Modifiers: m})
}
func (c *Connection) SendRune(t string) { c.input(protocol.RuneMsg{Type: protocol.TypeRune, Text: t}) }
