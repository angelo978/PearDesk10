package host

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/peardesk/peardesk/pkg/protocol"
	"github.com/peardesk/peardesk/pkg/tor"
	"net"
	"sync"
)

type Server struct {
	password string
	service  *tor.HostService
	cancel   context.CancelFunc
	OnLog    func(string)
}

func NewServer(password string) *Server { return &Server{password: password} }
func (s *Server) Onion() string {
	if s.service == nil {
		return ""
	}
	return s.service.Onion
}
func (s *Server) DataDir() string {
	if s.service == nil {
		return ""
	}
	return s.service.Dir
}
func (s *Server) Start(ctx context.Context, dataDir string) error {
	ctx, s.cancel = context.WithCancel(ctx)
	svc, err := tor.StartHost(ctx, dataDir)
	if err != nil {
		return err
	}
	s.service = svc
	s.logf("Tor Orion: %s", svc.Onion)
	go func() {
		for {
			c, e := svc.Listener().Accept()
			if e != nil {
				return
			}
			go s.handle(c)
		}
	}()
	return nil
}
func (s *Server) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.service != nil {
		s.service.Close()
	}
}
func (s *Server) handle(c net.Conn) {
	defer c.Close()
	typ, p, e := tor.ReadFrame(c)
	if e != nil || typ != tor.FrameAuth {
		return
	}
	var a struct {
		Password string `json:"password"`
	}
	_ = json.Unmarshal(p, &a)
	if s.password != "" && a.Password != s.password {
		_ = tor.WriteFrame(c, tor.FrameAuth, []byte(`{"ok":false}`))
		return
	}
	if tor.WriteFrame(c, tor.FrameAuth, []byte(`{"ok":true}`)) != nil {
		return
	}
	var mu sync.Mutex
	write := func(t byte, b []byte) error { mu.Lock(); defer mu.Unlock(); return tor.WriteFrame(c, t, b) }
	go s.videoLoop(write)
	go s.clipLoop(write)
	for {
		t, b, e := tor.ReadFrame(c)
		if e != nil {
			return
		}
		if t == tor.FrameInput {
			s.input(b)
		} else if t == tor.FrameClipboard {
			var x protocol.ClipboardMsg
			if json.Unmarshal(b, &x) == nil {
				applyRemoteClipboard(x.Text)
			}
		}
	}
}
func (s *Server) logf(f string, a ...interface{}) {
	if s.OnLog != nil {
		s.OnLog(fmt.Sprintf(f, a...))
	}
}
