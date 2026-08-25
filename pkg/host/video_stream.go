package host

import (
	"encoding/binary"
	"github.com/peardesk/peardesk/pkg/tor"
	"github.com/peardesk/peardesk/pkg/video"
	"time"
)

func (s *Server) videoLoop(write func(byte, []byte) error) {
	w, h := screenSize()
	if w == 0 {
		w, h = 1920, 1080
	}
	e, err := video.NewEncoder(w, h)
	if err != nil {
		return
	}
	defer e.Close()
	tick := time.NewTicker(33 * time.Millisecond)
	defer tick.Stop()
	for range tick.C {
		img, x, y, er := captureRaw(w, h)
		if er != nil || img == nil {
			continue
		}
		if x != w || y != h {
			e.Close()
			e, err = video.NewEncoder(x, y)
			if err != nil {
				return
			}
			w, h = x, y
		}
		nal, er := e.Encode(img)
		if er != nil || len(nal) == 0 {
			continue
		}
		b := make([]byte, 8)
		binary.BigEndian.PutUint32(b, uint32(w))
		binary.BigEndian.PutUint32(b[4:], uint32(h))
		if write(tor.FrameVideo, append(b, nal...)) != nil {
			return
		}
	}
}
