// Package tor is PearDesk-Tor's only transport. It contains no P2P, relay,
// WAN, Cloudflare, or fallback networking.
package tor

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	SOCKSAddr           = "127.0.0.1:9050"
	PortTCP             = 10001
	PortUDP             = 10002
	FrameAuth      byte = 'A'
	FrameVideo     byte = 'V'
	FrameInput     byte = 'I'
	FrameClipboard byte = 'C'
)

// WriteFrame and ReadFrame are the only wire framing functions.
func WriteFrame(c net.Conn, typ byte, payload []byte) error {
	var h [5]byte
	h[0] = typ
	binary.BigEndian.PutUint32(h[1:], uint32(len(payload)))
	if _, err := c.Write(h[:]); err != nil {
		return err
	}
	_, err := c.Write(payload)
	return err
}

func ReadFrame(c net.Conn) (byte, []byte, error) {
	var h [5]byte
	if _, err := io.ReadFull(c, h[:]); err != nil {
		return 0, nil, err
	}
	n := binary.BigEndian.Uint32(h[1:])
	if n > 64<<20 {
		return 0, nil, errors.New("Tor frame troppo grande")
	}
	p := make([]byte, n)
	_, err := io.ReadFull(c, p)
	return h[0], p, err
}

type HostService struct {
	Onion string
	Dir   string
	cmd   *exec.Cmd
	ln    net.Listener
}

// StartHost creates a v3 Onion service. The complete hostname in Onion is
// the only PearDesk host ID; it is never shortened or replaced.
func StartHost(ctx context.Context, persistentDir string) (*HostService, error) {
	root := persistentDir
	if root == "" {
		root, _ = os.MkdirTemp("", "peardesk-orion-")
	}
	hidden := filepath.Join(root, "hidden")
	if err := os.MkdirAll(hidden, 0700); err != nil {
		return nil, err
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	rc := fmt.Sprintf("DataDirectory %s\nSocksPort 0\nHiddenServiceDir %s\nHiddenServicePort %d 127.0.0.1:%d\n", filepath.Join(root, "data"), hidden, PortTCP, ln.Addr().(*net.TCPAddr).Port)
	rcPath := filepath.Join(root, "torrc")
	if err := os.WriteFile(rcPath, []byte(rc), 0600); err != nil {
		ln.Close()
		return nil, err
	}
	cmd := exec.CommandContext(ctx, "tor", "-f", rcPath)
	if err := cmd.Start(); err != nil {
		ln.Close()
		return nil, fmt.Errorf("Tor Orion non installato: %w", err)
	}
	hostname := filepath.Join(hidden, "hostname")
	var onion string
	for deadline := time.Now().Add(60 * time.Second); time.Now().Before(deadline); {
		if b, e := os.ReadFile(hostname); e == nil {
			onion = strings.TrimSpace(string(b))
			if onion != "" {
				break
			}
		}
		select {
		case <-ctx.Done():
			_ = cmd.Process.Kill()
			ln.Close()
			return nil, ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
	if onion == "" {
		_ = cmd.Process.Kill()
		ln.Close()
		return nil, errors.New("Tor Orion non ha generato l'Onion address")
	}
	return &HostService{Onion: onion, Dir: root, cmd: cmd, ln: ln}, nil
}

func (s *HostService) Listener() net.Listener { return s.ln }
func (s *HostService) Close() {
	if s.ln != nil {
		s.ln.Close()
	}
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	// Keep the persistent hidden-service directory: it contains the Onion
	// private key and is what makes the hostname stable across restarts.
}

// Dial connects only through the local Tor SOCKS5 proxy.
func Dial(ctx context.Context, onion string) (net.Conn, error) {
	d := net.Dialer{Timeout: 30 * time.Second}
	c, err := d.DialContext(ctx, "tcp", SOCKSAddr)
	if err != nil {
		return nil, fmt.Errorf("Tor SOCKS5 non disponibile su %s: %w", SOCKSAddr, err)
	}
	if _, err = c.Write([]byte{5, 1, 0}); err != nil {
		c.Close()
		return nil, err
	}
	var hello [2]byte
	if _, err = io.ReadFull(c, hello[:]); err != nil || hello != [2]byte{5, 0} {
		c.Close()
		return nil, errors.New("proxy Tor SOCKS5 rifiuta la connessione")
	}
	host := []byte(onion)
	if len(host) != 62 || !strings.HasSuffix(onion, ".onion") {
		c.Close()
		return nil, errors.New("ID Onion completo non valido")
	}
	req := []byte{5, 1, 0, 3, byte(len(host))}
	req = append(req, host...)
	var port [2]byte
	binary.BigEndian.PutUint16(port[:], PortTCP)
	req = append(req, port[:]...)
	if _, err = c.Write(req); err != nil {
		c.Close()
		return nil, err
	}
	var h [4]byte
	if _, err = io.ReadFull(c, h[:]); err != nil {
		c.Close()
		return nil, err
	}
	if h[1] != 0 {
		c.Close()
		return nil, fmt.Errorf("Tor SOCKS5 errore %d", h[1])
	}
	var n int
	switch h[3] {
	case 1:
		n = 4
	case 4:
		n = 16
	case 3:
		var l [1]byte
		if _, err = io.ReadFull(c, l[:]); err != nil {
			c.Close()
			return nil, err
		}
		n = int(l[0])
	default:
		c.Close()
		return nil, errors.New("risposta Tor SOCKS5 non valida")
	}
	buf := make([]byte, n+2)
	if _, err = io.ReadFull(c, buf); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}
