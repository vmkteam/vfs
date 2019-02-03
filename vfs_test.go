package vfs_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"vfs"
)

func ExampleFileHash() {
	fh := vfs.FileHash("6698364ea6730f327a26bb8a6d3da3be")
	fmt.Println(fh.Dir())
	fmt.Println(fh.File())
	// Output:
	// 6/69
	// 6/69/6698364ea6730f327a26bb8a6d3da3be.jpg
}

func TestVFS_Upload(t *testing.T) {
	err := os.MkdirAll("testdata", os.ModePerm)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	v, err := vfs.New(vfs.Config{Path: "testdata"})
	if err != nil {
		t.Fatalf("failed to create vfs: %v", err)
	}

	// hash upload
	data := strings.Repeat("temp file", 100)
	path, err := v.HashUpload(strings.NewReader(data), vfs.NamespacePublic)
	if err != nil {
		t.Errorf("failed to perform hash upload: %v", err)
	} else {
		t.Logf("hash=%v dir=%s file=%s", path, path.Dir(), path.File())
	}

	err = v.Upload(strings.NewReader(data), "201901/123_456.png", vfs.NamespacePublic)
	if err != nil {
		t.Errorf("failed to perform upload: %v", err)
	}
}
