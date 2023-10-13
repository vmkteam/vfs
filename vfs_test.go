package vfs_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/vmkteam/vfs"
)

func ExampleFileHash() {
	fh := vfs.NewFileHash("6698364ea6730f327a26bb8a6d3da3be", "")
	fmt.Println(fh.Dir())
	fmt.Println(fh.File())

	// automatic rewrite jpeg extension to jpg
	fh2 := vfs.NewFileHash("6698364ea6730f327a26bb8a6d3da3be", "jpeg")
	fmt.Println(fh2.File())

	// Output:
	// 6/69
	// 6/69/6698364ea6730f327a26bb8a6d3da3be.jpg
	// 6/69/6698364ea6730f327a26bb8a6d3da3be.jpg
}

func TestVFS_Upload(t *testing.T) {
	err := os.MkdirAll("testdata", os.ModePerm)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	v, err := vfs.New(vfs.Config{Path: "testdata", Extensions: []string{"png"}, MimeTypes: []string{"image/png"}})
	if err != nil {
		t.Fatalf("failed to create vfs: %v", err)
	}

	// hash upload
	data, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=")
	if err != nil {
		t.Errorf("failed to decode base64 image: %v", err)
	}
	path, err := v.HashUpload(bytes.NewReader(data), vfs.NamespacePublic, "png")
	if err != nil {
		t.Errorf("failed to perform hash upload: %v", err)
	} else {
		t.Logf("hash=%v dir=%s file=%s", path, path.Dir(), path.File())
	}

	err = v.Upload(bytes.NewReader(data), "201901/123_456.png", vfs.NamespacePublic)
	if err != nil {
		t.Errorf("failed to perform upload: %v", err)
	}
}
