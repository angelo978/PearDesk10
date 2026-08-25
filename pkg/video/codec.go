//go:build !headless

// Package video provides H264 encoding (screen→NAL) and decoding (NAL→image)
// using FFmpeg's libavcodec via CGO.
package video

/*
#cgo pkg-config: libavcodec libavutil libswscale
#include <libavcodec/avcodec.h>
#include <libavutil/error.h>
#include <libavutil/imgutils.h>
#include <libswscale/swscale.h>
#include <errno.h>
#include <stdlib.h>
#include <string.h>

static const int AV_PADDING = AV_INPUT_BUFFER_PADDING_SIZE;

// averror_eagain / averror_eof wrap the AVERROR macro so CGO can use them.
static int averror_eagain(void) { return AVERROR(EAGAIN); }
static int averror_eof(void)    { return AVERROR_EOF; }

// alloc_yuv_frame allocates an AVFrame configured for YUV420P.
static AVFrame* alloc_yuv_frame(int width, int height) {
    AVFrame *f = av_frame_alloc();
    if (!f) return NULL;
    f->format = AV_PIX_FMT_YUV420P;
    f->width  = width;
    f->height = height;
    if (av_frame_get_buffer(f, 0) < 0) {
        av_frame_free(&f);
        return NULL;
    }
    return f;
}

// make_packet_from_data creates an AVPacket owning a copy of the provided data.
static AVPacket* make_packet_from_data(const uint8_t *src, int size) {
    AVPacket *pkt = av_packet_alloc();
    if (!pkt) return NULL;
    int pad = AV_INPUT_BUFFER_PADDING_SIZE;
    uint8_t *buf = (uint8_t*)av_malloc(size + pad);
    if (!buf) { av_packet_free(&pkt); return NULL; }
    memcpy(buf, src, size);
    memset(buf + size, 0, pad);
    av_packet_from_data(pkt, buf, size);
    return pkt;
}
*/
import "C"
import (
	"fmt"
	"image"
	"unsafe"
)

// ─── Encoder ─────────────────────────────────────────────────────────────────

// Encoder encodes raw screen images to H264 NAL units using libx264.
type Encoder struct {
	codec  *C.AVCodec
	ctx    *C.AVCodecContext
	frame  *C.AVFrame
	pkt    *C.AVPacket
	Width  int
	Height int
	pts    C.int64_t
}

// NewEncoder creates a new H264 encoder for frames of the given size.
func NewEncoder(width, height int) (*Encoder, error) {
	codec := C.avcodec_find_encoder(C.AV_CODEC_ID_H264)
	if codec == nil {
		return nil, fmt.Errorf("H264 encoder not found (libx264 missing?)")
	}

	ctx := C.avcodec_alloc_context3(codec)
	if ctx == nil {
		return nil, fmt.Errorf("avcodec_alloc_context3 failed")
	}

	ctx.bit_rate = 4_000_000
	ctx.width = C.int(width)
	ctx.height = C.int(height)
	ctx.time_base.num = 1
	ctx.time_base.den = 30
	ctx.framerate.num = 30
	ctx.framerate.den = 1
	ctx.pix_fmt = C.AV_PIX_FMT_YUV420P
	ctx.max_b_frames = 0
	ctx.gop_size = 30

	// Low-latency preset via AVDictionary
	var opts *C.AVDictionary
	setDictEntry(&opts, "preset", "veryfast")
	setDictEntry(&opts, "tune", "zerolatency")
	setDictEntry(&opts, "x264-params", "keyint=30:min-keyint=30")
	defer C.av_dict_free(&opts)

	if ret := C.avcodec_open2(ctx, codec, &opts); ret < 0 {
		C.avcodec_free_context(&ctx)
		return nil, fmt.Errorf("avcodec_open2 failed: %d", int(ret))
	}

	frame := C.alloc_yuv_frame(C.int(width), C.int(height))
	if frame == nil {
		C.avcodec_free_context(&ctx)
		return nil, fmt.Errorf("alloc_yuv_frame failed")
	}

	return &Encoder{
		codec:  codec,
		ctx:    ctx,
		frame:  frame,
		pkt:    C.av_packet_alloc(),
		Width:  width,
		Height: height,
	}, nil
}

// Encode converts img to a H264 NAL unit. Returns nil, nil if the encoder
// needs more frames before producing output (normal for the first few frames).
func (e *Encoder) Encode(img image.Image) ([]byte, error) {
	if ret := C.av_frame_make_writable(e.frame); ret < 0 {
		return nil, fmt.Errorf("av_frame_make_writable: %d", int(ret))
	}

	imageToYUV420P(img, e.frame)
	e.frame.pts = e.pts
	e.pts++

	if ret := C.avcodec_send_frame(e.ctx, e.frame); ret < 0 {
		return nil, fmt.Errorf("avcodec_send_frame: %d", int(ret))
	}

	C.av_packet_unref(e.pkt)
	ret := C.avcodec_receive_packet(e.ctx, e.pkt)
	if ret == C.int(C.averror_eagain()) || ret == C.int(C.averror_eof()) {
		return nil, nil // need more frames
	}
	if ret < 0 {
		return nil, fmt.Errorf("avcodec_receive_packet: %d", int(ret))
	}

	out := C.GoBytes(unsafe.Pointer(e.pkt.data), e.pkt.size)
	return out, nil
}

// Close frees all encoder resources.
func (e *Encoder) Close() {
	if e.pkt != nil {
		C.av_packet_free(&e.pkt)
	}
	if e.frame != nil {
		C.av_frame_free(&e.frame)
	}
	if e.ctx != nil {
		C.avcodec_free_context(&e.ctx)
	}
}

// ─── Decoder ─────────────────────────────────────────────────────────────────

// Decoder decodes H264 NAL units to image.RGBA using libavcodec.
type Decoder struct {
	ctx   *C.AVCodecContext
	frame *C.AVFrame
}

// NewDecoder creates a new H264 software decoder.
func NewDecoder() (*Decoder, error) {
	codec := C.avcodec_find_decoder(C.AV_CODEC_ID_H264)
	if codec == nil {
		return nil, fmt.Errorf("H264 decoder not found")
	}

	ctx := C.avcodec_alloc_context3(codec)
	if ctx == nil {
		return nil, fmt.Errorf("avcodec_alloc_context3 failed")
	}

	if ret := C.avcodec_open2(ctx, codec, nil); ret < 0 {
		C.avcodec_free_context(&ctx)
		return nil, fmt.Errorf("avcodec_open2 failed: %d", int(ret))
	}

	return &Decoder{
		ctx:   ctx,
		frame: C.av_frame_alloc(),
	}, nil
}

// Decode decodes a H264 NAL unit to an image. Returns nil, nil if the decoder
// needs more NAL units before producing a full frame (buffering).
func (d *Decoder) Decode(nalData []byte) (image.Image, error) {
	if len(nalData) == 0 {
		return nil, nil
	}

	pkt := C.make_packet_from_data(
		(*C.uint8_t)(unsafe.Pointer(&nalData[0])),
		C.int(len(nalData)),
	)
	if pkt == nil {
		return nil, fmt.Errorf("make_packet_from_data failed")
	}
	defer C.av_packet_free(&pkt)

	if ret := C.avcodec_send_packet(d.ctx, pkt); ret < 0 {
		return nil, fmt.Errorf("avcodec_send_packet: %d", int(ret))
	}

	C.av_frame_unref(d.frame)
	ret := C.avcodec_receive_frame(d.ctx, d.frame)
	if ret == C.int(C.averror_eagain()) || ret == C.int(C.averror_eof()) {
		return nil, nil // need more packets
	}
	if ret < 0 {
		return nil, fmt.Errorf("avcodec_receive_frame: %d", int(ret))
	}

	return yuv420PToImage(d.frame), nil
}

// Close frees all decoder resources.
func (d *Decoder) Close() {
	if d.frame != nil {
		C.av_frame_free(&d.frame)
	}
	if d.ctx != nil {
		C.avcodec_free_context(&d.ctx)
	}
}

// ─── YUV ↔ image conversion ──────────────────────────────────────────────────

// imageToYUV420P fills a YUV420P AVFrame from a Go image.Image.
func imageToYUV420P(img image.Image, frame *C.AVFrame) {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	minX, minY := bounds.Min.X, bounds.Min.Y

	yStride := int(frame.linesize[0])
	uStride := int(frame.linesize[1])

	yPlane := unsafe.Slice((*byte)(unsafe.Pointer(frame.data[0])), yStride*h)
	uPlane := unsafe.Slice((*byte)(unsafe.Pointer(frame.data[1])), uStride*((h+1)/2))
	vPlane := unsafe.Slice((*byte)(unsafe.Pointer(frame.data[2])), uStride*((h+1)/2))

	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			r, g, b, _ := img.At(col+minX, row+minY).RGBA()
			ri, gi, bi := int(r>>8), int(g>>8), int(b>>8)

			// BT.601 limited range
			yVal := (66*ri+129*gi+25*bi+128)>>8 + 16
			yPlane[row*yStride+col] = uint8(clamp8(yVal))

			if row%2 == 0 && col%2 == 0 {
				uVal := (-38*ri-74*gi+112*bi+128)>>8 + 128
				vVal := (112*ri-94*gi-18*bi+128)>>8 + 128
				uPlane[(row/2)*uStride+(col/2)] = uint8(clamp8(uVal))
				vPlane[(row/2)*uStride+(col/2)] = uint8(clamp8(vVal))
			}
		}
	}
}

// yuv420PToImage converts a decoded YUV420P AVFrame to image.RGBA.
func yuv420PToImage(frame *C.AVFrame) *image.RGBA {
	w := int(frame.width)
	h := int(frame.height)

	yStride := int(frame.linesize[0])
	uStride := int(frame.linesize[1])

	yPlane := unsafe.Slice((*byte)(unsafe.Pointer(frame.data[0])), yStride*h)
	uPlane := unsafe.Slice((*byte)(unsafe.Pointer(frame.data[1])), uStride*((h+1)/2))
	vPlane := unsafe.Slice((*byte)(unsafe.Pointer(frame.data[2])), uStride*((h+1)/2))

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	for row := 0; row < h; row++ {
		base := row * img.Stride
		for col := 0; col < w; col++ {
			// BT.601 limited range → full range
			y := int(yPlane[row*yStride+col]) - 16
			u := int(uPlane[(row/2)*uStride+(col/2)]) - 128
			v := int(vPlane[(row/2)*uStride+(col/2)]) - 128

			r := clamp8((298*y + 409*v + 128) >> 8)
			g := clamp8((298*y - 100*u - 208*v + 128) >> 8)
			b := clamp8((298*y + 516*u + 128) >> 8)

			idx := base + col*4
			img.Pix[idx] = uint8(r)
			img.Pix[idx+1] = uint8(g)
			img.Pix[idx+2] = uint8(b)
			img.Pix[idx+3] = 255
		}
	}
	return img
}

func clamp8(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// setDictEntry sets a key/value pair in an AVDictionary.
func setDictEntry(d **C.AVDictionary, key, val string) {
	k := C.CString(key)
	v := C.CString(val)
	C.av_dict_set(d, k, v, 0)
	C.free(unsafe.Pointer(k))
	C.free(unsafe.Pointer(v))
}
