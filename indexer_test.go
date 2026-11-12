package vfs

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/vmkteam/embedlog"
	"github.com/vmkteam/vfs/db"
)

func Test_isHashFile(t *testing.T) {
	type args struct {
		ns   string
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty ns & path",
			args: args{ns: "", path: ""},
			want: false,
		},
		{
			name: "empty path",
			args: args{ns: "test-namespace", path: ""},
			want: false,
		},
		{
			name: "wrong length",
			args: args{ns: "test-namespace", path: "test-namespace/test-file.jpeg"},
			want: false,
		},
		{
			name: "wrong dir tree",
			args: args{ns: "", path: "70/c/70c565ef460af43688b7ee6251028db9.jpg"},
			want: false,
		},
		{
			name: "wrong dir tree with namespace",
			args: args{ns: "test", path: "test/70/c/70c565ef460af43688b7ee6251028db9.jpg"},
			want: false,
		},
		{
			name: "wrong hash prefix",
			args: args{ns: "test", path: "test/7/1c/70c565ef460af43688b7ee6251028db9.jpg"},
			want: false,
		},
		{
			name: "non-hex path",
			args: args{ns: "", path: "7/0q/70q565ef460af43688b7ee6251028db9.jpg"},
			want: false,
		},
		{
			name: "ok without extension",
			args: args{ns: "test", path: "test/7/0f/70f565ef460af43688b7ee6251028db9"},
			want: true,
		},
		{
			name: "ok with 3-chars extension",
			args: args{ns: "", path: "7/0c/70c565ef460af43688b7ee6251028db9.gif"},
			want: true,
		},
		{
			name: "ok with 4-chars extension",
			args: args{ns: "", path: "7/0c/70c565ef460af43688b7ee6251028db9.jpeg"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isHashFile(tt.args.ns, tt.args.path); got != tt.want {
				t.Errorf("isHashFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

// writePNG writes a picture of w×h, or only the png header when header is set: enough for image.DecodeConfig.
func writePNG(t *testing.T, file string, w, h int, header bool) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if header {
		ihdr := binary.BigEndian.AppendUint32([]byte("IHDR"), uint32(w))
		ihdr = binary.BigEndian.AppendUint32(ihdr, uint32(h))
		ihdr = append(ihdr, 8, 0, 0, 0, 0)
		out := binary.BigEndian.AppendUint32([]byte("\x89PNG\r\n\x1a\n"), uint32(len(ihdr)-4))
		out = binary.BigEndian.AppendUint32(append(out, ihdr...), crc32.ChecksumIEEE(ihdr))
		if err := os.WriteFile(file, out, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHashIndexer_IndexFile(t *testing.T) {
	root := t.TempDir()
	v, err := New(Config{Path: root}, embedlog.NewDevLogger())
	if err != nil {
		t.Fatal(err)
	}
	hi := NewHashIndexer(embedlog.NewDevLogger(), db.DB{}, nil, v, 1, 1, true, nil)

	t.Run("a picture gets its size and a blurhash", func(t *testing.T) {
		writePNG(t, filepath.Join(root, "ns", "a.png"), 300, 200, false)
		info, err := hi.IndexFile("ns", "a.png")
		if err != nil {
			t.Fatal(err)
		}
		if info.Width != 300 || info.Height != 200 || info.BlurHash == "" || info.FileSize == 0 {
			t.Errorf("got %+v", info)
		}
	})

	t.Run("a huge picture gets its size only and is not decoded", func(t *testing.T) {
		// 6000×5000 is over MaxBlurHashPixels; the file has no pixel data, a decode would fail
		writePNG(t, filepath.Join(root, "ns", "huge.png"), 6000, 5000, true)
		info, err := hi.IndexFile("ns", "huge.png")
		if err != nil {
			t.Fatal(err)
		}
		if info.Width != 6000 || info.Height != 5000 || info.BlurHash != "" {
			t.Errorf("got %+v", info)
		}
	})

	t.Run("without blurhash a picture is not decoded", func(t *testing.T) {
		writePNG(t, filepath.Join(root, "ns", "header.png"), 300, 200, true)
		noBH := NewHashIndexer(embedlog.NewDevLogger(), db.DB{}, nil, v, 1, 1, false, nil)
		info, err := noBH.IndexFile("ns", "header.png")
		if err != nil {
			t.Fatal(err)
		}
		if info.Width != 300 || info.BlurHash != "" {
			t.Errorf("got %+v", info)
		}
	})
}
