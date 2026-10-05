// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/io/transfer"
)

func pasteData(typ, content string) transfer.DataEvent {
	return transfer.DataEvent{Type: typ, Open: func() io.ReadCloser { return io.NopCloser(strings.NewReader(content)) }}
}

// Ctrl+V in the message field reads files, then a picture, then text from
// the clipboard; files and pictures open the box for sending files.
func TestComposerPaste(t *testing.T) {
	h := newComposerHarness(t)
	c := h.p.composer
	h.click(150, 680)
	paste := func(event transfer.DataEvent) {
		t.Helper()
		h.router.Queue(key.Event{Name: "V", Modifiers: key.ModShortcut, State: key.Press})
		h.frame()
		types, ok := h.router.ClipboardRequested()
		if want := []string{clipboard.TypeURIList, clipboard.TypePNG, clipboard.TypeText}; !ok || !slices.Equal(types, want) {
			t.Fatalf("Ctrl+V read %v, %v", types, ok)
		}
		h.router.Queue(event)
		h.frame()
	}

	paste(pasteData(clipboard.TypeText, "hello"))
	if got := c.draft(1).editor.Text(); got != "hello" || c.draft(1).text != "hello" {
		t.Fatalf("pasted text %q", got)
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "a b.txt")
	os.WriteFile(file, []byte("x"), 0o600)
	paste(pasteData(clipboard.TypeURIList, (&url.URL{Scheme: "file", Path: file}).String()+"\r\n"))
	if !c.files.Shown() || len(c.files.files) != 1 || c.files.files[0].Path != file {
		t.Fatalf("pasted files: shown %v, %d", c.files.Shown(), len(c.files.files))
	}
	if got := c.files.caption.Text(); got != "hello" {
		t.Fatalf("caption %q", got)
	}
	c.files.close()
	h.frame()
	h.click(150, 680)

	// Links are not files: the picture or the text is read instead.
	h.router.Queue(key.Event{Name: "V", Modifiers: key.ModShortcut, State: key.Press})
	h.frame()
	h.router.ClipboardRequested()
	h.router.Queue(pasteData(clipboard.TypeURIList, "https://telegram.org\r\n"))
	h.frame()
	if types, ok := h.router.ClipboardRequested(); !ok || !slices.Equal(types, []string{clipboard.TypePNG, clipboard.TypeText}) {
		t.Fatalf("after links read %v, %v", types, ok)
	}
	h.router.Queue(pasteData(clipboard.TypeText, ""))
	h.frame()
	h.click(150, 680)

	var pic bytes.Buffer
	png.Encode(&pic, image.NewNRGBA(image.Rect(0, 0, 4, 3)))
	paste(pasteData(clipboard.TypePNG, pic.String()))
	if !c.files.Shown() || len(c.files.files) != 1 || filepath.Base(c.files.files[0].Path) != "image.png" {
		t.Fatalf("pasted picture: shown %v, %d", c.files.Shown(), len(c.files.files))
	}
	saved := c.files.files[0].Path
	if b, err := os.ReadFile(saved); err != nil || !bytes.Equal(b, pic.Bytes()) {
		t.Fatalf("saved picture: %v", err)
	}
	c.files.modal.Hide()
	h.frame()
	if _, err := os.Stat(saved); !os.IsNotExist(err) {
		t.Fatalf("a picture given up was kept: %v", err)
	}
}

func TestLocalPaths(t *testing.T) {
	if got := localPaths("# copied\r\nfile:///tmp/a%20b.png\r\nfile://localhost/tmp/c\r\n"); !slices.Equal(got, []string{"/tmp/a b.png", "/tmp/c"}) {
		t.Fatalf("got %q", got)
	}
	if got := localPaths("file:///tmp/a\r\nhttps://telegram.org\r\n"); got != nil {
		t.Fatalf("a list with a link gave %q", got)
	}
}
