package ui

import (
	"fmt"
	"image"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/peardesk/peardesk/pkg/client"
	"github.com/peardesk/peardesk/pkg/i18n"
)

// RemoteWindow displays the remote desktop feed and routes input events.
type RemoteWindow struct {
	win       fyne.Window
	conn      *client.Connection
	imgWidget *canvas.Image
	statusLbl *widget.Label
	cbLbl     *widget.Label
	mu        sync.Mutex
	remoteW   int
	remoteH   int
}

func ShowRemoteWindow(a fyne.App, conn *client.Connection, hostID string) *RemoteWindow {
	win := a.NewWindow("PearDesk — " + hostID)
	win.Resize(fyne.NewSize(1280, 720))
	win.SetFixedSize(false)

	rw := &RemoteWindow{
		win:       win,
		conn:      conn,
		statusLbl: widget.NewLabel(i18n.T("connected_to") + " " + hostID),
		cbLbl:     widget.NewLabel(""),
	}

	img := canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1280, 720)))
	img.FillMode = canvas.ImageFillContain
	img.ScaleMode = canvas.ImageScaleFastest
	rw.imgWidget = img

	interactiveImg := newInteractiveImage(img, rw)

	toolbar := container.NewBorder(nil, nil, rw.statusLbl, rw.cbLbl, nil)
	content := container.NewBorder(toolbar, nil, nil, nil, interactiveImg)
	win.SetContent(content)

	// ── Keyboard: typed rune (uppercase, @, accented chars) ───────────────────
	win.Canvas().SetOnTypedRune(func(r rune) {
		conn.SendRune(string(r))
	})

	// ── Keyboard: special keys (F1-F12, arrows, ctrl, alt, tab, esc…) ────────
	win.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if isPrintableKey(string(ev.Name)) {
			return // handled by OnTypedRune
		}
		conn.SendKeyDown(string(ev.Name), nil)
		conn.SendKeyUp(string(ev.Name), nil)
	})

	// ── Video feed ────────────────────────────────────────────────────────────
	conn.OnFrame = func(img image.Image) { rw.updateFrame(img) }
	conn.OnClose = func() { rw.statusLbl.SetText(i18n.T("connection_closed")) }
	conn.OnError = func(err error) {
		rw.statusLbl.SetText(i18n.T("error") + ": " + err.Error())
	}

	// ── Clipboard feedback ────────────────────────────────────────────────────
	conn.OnClipboard = func(_ string) { rw.flashClipboard("← host") }

	win.SetOnClosed(func() { conn.Close() })
	win.Show()
	return rw
}

func (rw *RemoteWindow) updateFrame(img image.Image) {
	bounds := img.Bounds()
	rw.mu.Lock()
	rw.remoteW = bounds.Dx()
	rw.remoteH = bounds.Dy()
	rw.mu.Unlock()
	rw.imgWidget.Image = img
	rw.imgWidget.Refresh()
}

func (rw *RemoteWindow) flashClipboard(direction string) {
	rw.cbLbl.SetText(fmt.Sprintf("📋 %s", direction))
	go func() {
		time.Sleep(2 * time.Second)
		rw.cbLbl.SetText("")
	}()
}

// imageRect returns the rendered image area within the widget,
// accounting for letterboxing from ImageFillContain.
func (rw *RemoteWindow) imageRect() (offsetX, offsetY, imgW, imgH float32) {
	rw.mu.Lock()
	remW := rw.remoteW
	remH := rw.remoteH
	rw.mu.Unlock()

	ws := rw.imgWidget.Size()
	wW, wH := ws.Width, ws.Height

	if remW == 0 || remH == 0 || wW == 0 || wH == 0 {
		return 0, 0, wW, wH
	}

	scaleX := wW / float32(remW)
	scaleY := wH / float32(remH)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}

	imgW = float32(remW) * scale
	imgH = float32(remH) * scale
	offsetX = (wW - imgW) / 2
	offsetY = (wH - imgH) / 2
	return
}

func isPrintableKey(key string) bool {
	if len(key) == 1 {
		r := rune(key[0])
		return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r > 127
	}
	switch key {
	case "Space", "Return", "KP_Enter":
		return true
	}
	return false
}

// ─── Interactive image widget ─────────────────────────────────────────────────

type interactiveImage struct {
	widget.BaseWidget
	img *canvas.Image
	rw  *RemoteWindow
}

func newInteractiveImage(img *canvas.Image, rw *RemoteWindow) *interactiveImage {
	i := &interactiveImage{img: img, rw: rw}
	i.ExtendBaseWidget(i)
	return i
}

func (i *interactiveImage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(i.img)
}

func (i *interactiveImage) MouseIn(_ *desktop.MouseEvent) {}
func (i *interactiveImage) MouseOut()                     {}
func (i *interactiveImage) MouseMoved(ev *desktop.MouseEvent) {
	xR, yR := i.ratios(ev.Position)
	i.rw.conn.SendMouseMove(xR, yR)
}
func (i *interactiveImage) Tapped(ev *fyne.PointEvent) {
	xR, yR := i.ratios(ev.Position)
	i.rw.conn.SendMouseDown(xR, yR, "left")
	i.rw.conn.SendMouseUp(xR, yR, "left")
}
func (i *interactiveImage) TappedSecondary(ev *fyne.PointEvent) {
	xR, yR := i.ratios(ev.Position)
	i.rw.conn.SendMouseDown(xR, yR, "right")
	i.rw.conn.SendMouseUp(xR, yR, "right")
}
func (i *interactiveImage) Scrolled(ev *fyne.ScrollEvent) {
	xR, yR := i.ratios(ev.Position)
	i.rw.conn.SendScroll(xR, yR, float64(ev.Scrolled.DY))
}

// ratios converts a widget-local position to [0,1] ratios on the remote screen,
// correcting for letterbox offsets from ImageFillContain.
func (i *interactiveImage) ratios(pos fyne.Position) (float64, float64) {
	offsetX, offsetY, imgW, imgH := i.rw.imageRect()
	if imgW == 0 || imgH == 0 {
		return 0, 0
	}
	x := float64(pos.X-offsetX) / float64(imgW)
	y := float64(pos.Y-offsetY) / float64(imgH)
	if x < 0 {
		x = 0
	} else if x > 1 {
		x = 1
	}
	if y < 0 {
		y = 0
	} else if y > 1 {
		y = 1
	}
	return x, y
}
