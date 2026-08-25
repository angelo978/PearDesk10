// Package protocol defines the JSON message types exchanged over the PearDesk
// WAN TCP channel (port 4663). UDP port 4673 carries raw mouse-move JSON only.
package protocol

// Message type discriminators.
const (
	TypeAuth       = "auth"
	TypeAuthOK     = "auth_ok"
	TypeAuthFail   = "auth_fail"
	TypeMouseEvent = "mouse"
	TypeKeyEvent   = "key"
	TypeRune       = "rune"      // direct Unicode character injection
	TypeClipboard  = "clipboard" // bidirectional clipboard sync

	// File transfer
	TypeFileListReq      = "file_list_req"
	TypeFileList         = "file_list"
	TypeFileDownloadReq  = "file_download_req"
	TypeFileChunk        = "file_chunk"
	TypeFileTransferDone = "file_transfer_done"
	TypeFileTransferErr  = "file_transfer_err"
	TypeFileUploadStart  = "file_upload_start"
	TypeFileUploadReady  = "file_upload_ready"
	TypeFileUploadDone   = "file_upload_done"
)

// Message is used to peek at the "type" field before full unmarshal.
type Message struct {
	Type string `json:"type"`
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

type AuthMsg struct {
	Type     string `json:"type"`
	Password string `json:"password"`
}

type AuthResultMsg struct {
	Type  string `json:"type"`
	Token string `json:"token,omitempty"` // session token returned on auth_ok
}

// ─── Input ────────────────────────────────────────────────────────────────────

// MouseEventMsg is sent for every mouse action.
// X and Y are normalised coordinates in [0,1] relative to the remote screen.
type MouseEventMsg struct {
	Type    string  `json:"type"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Button  string  `json:"button"`             // "left"|"right"|"middle"|"none"
	Action  string  `json:"action"`             // "move"|"down"|"up"|"scroll"
	ScrollY float64 `json:"scroll_y,omitempty"` // positive = down
}

// KeyEventMsg is sent for physical key press / release events.
type KeyEventMsg struct {
	Type      string   `json:"type"`
	Key       string   `json:"key"`
	Action    string   `json:"action"`              // "down"|"up"
	Modifiers []string `json:"modifiers,omitempty"` // e.g. ["shift","ctrl"]
}

// RuneMsg carries a Unicode character for direct text injection.
type RuneMsg struct {
	Type string `json:"type"`
	Text string `json:"text"` // UTF-8 character(s)
}

// ─── Clipboard ────────────────────────────────────────────────────────────────

// ClipboardMsg is exchanged in both directions when clipboard content changes.
type ClipboardMsg struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ─── File browsing ────────────────────────────────────────────────────────────

type FileListReqMsg struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

type FileInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"is_dir"`
	ModTime string `json:"mod_time"`
}

type FileListMsg struct {
	Type  string     `json:"type"`
	Path  string     `json:"path"`
	Files []FileInfo `json:"files"`
}

// ─── Download (host → client) ─────────────────────────────────────────────────

type FileDownloadReqMsg struct {
	Type       string `json:"type"`
	TransferID string `json:"transfer_id"`
	Path       string `json:"path"` // full remote path
}

type FileChunkMsg struct {
	Type       string `json:"type"`
	TransferID string `json:"transfer_id"`
	Name       string `json:"name"`
	Index      int64  `json:"index"`
	Total      int64  `json:"total"`
	FileSize   int64  `json:"file_size"`
	Data       string `json:"data"` // base64-encoded chunk bytes
}

type FileTransferDoneMsg struct {
	Type       string `json:"type"`
	TransferID string `json:"transfer_id"`
	Name       string `json:"name"`
}

type FileTransferErrMsg struct {
	Type       string `json:"type"`
	TransferID string `json:"transfer_id"`
	Error      string `json:"error"`
}

// ─── Upload (client → host) ───────────────────────────────────────────────────

type FileUploadStartMsg struct {
	Type       string `json:"type"`
	TransferID string `json:"transfer_id"`
	Name       string `json:"name"`
	FileSize   int64  `json:"file_size"`
	Total      int64  `json:"total"`     // total chunks
	DestPath   string `json:"dest_path"` // remote destination directory
}

type FileUploadReadyMsg struct {
	Type       string `json:"type"`
	TransferID string `json:"transfer_id"`
}

type FileUploadDoneMsg struct {
	Type       string `json:"type"`
	TransferID string `json:"transfer_id"`
	SavedPath  string `json:"saved_path"`
}
