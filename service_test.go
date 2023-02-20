package vfs_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"os"
	"testing"

	"github.com/vmkteam/vfs"
	"github.com/vmkteam/vfs/db"

	"github.com/go-pg/pg/v10"
)

var (
	dbConn  = flag.String("db.conn", "postgresql://localhost:5432/vfs?sslmode=disable", "database connection dsn")
	service vfs.Service
	testVfs vfs.VFS
)

const testNs = "testns"

func TestMain(m *testing.M) {
	flag.Parse()
	cfg, err := pg.ParseURL(*dbConn)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll("testdata", os.ModePerm)
	if err != nil {
		panic(err)
	}

	v, err := vfs.New(vfs.Config{Path: "testdata", Extensions: []string{"png"}, MimeTypes: []string{"image/png"}, Namespaces: []string{testNs}})
	if err != nil {
		panic(err)
	}
	testVfs = v

	dbc := pg.Connect(cfg)
	repo := db.NewVfsRepo(db.New(dbc))
	service = vfs.NewService(repo, testVfs, dbc)
	os.Exit(m.Run())
}

func TestDBService_GetFolder(t *testing.T) {
	ctx := context.Background()

	// get folder
	folders, err := service.GetFolder(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	d, _ := json.Marshal(folders)
	if string(d) != `{"id":1,"name":"root","parentId":null,"folders":[{"id":2,"name":"test","parentId":1}]}` {
		t.Fatal(string(d))
	}

	// get branch
	branch, err := service.GetFolderBranch(ctx, 3)
	if err != nil {
		t.Fatal(err)
	}

	d, _ = json.Marshal(branch)
	if string(d) != `[{"id":1,"name":"root","parentId":null},{"id":2,"name":"test","parentId":1},{"id":3,"name":"test2","parentId":2}]` {
		t.Fatal(string(d))
	}
}

func TestDBService_GetFiles(t *testing.T) {
	ctx := context.Background()

	// get files
	q := "photo"
	files, err := service.GetFiles(ctx, 1, &q, "createdAt", true, 0, 100)
	if err != nil {
		t.Fatal(err)
	}

	d, _ := json.Marshal(files)
	if string(d) != `[{"id":9,"name":"photo_2019-07-30_14-18-07","path":"201908/1_9_ab990f98.jpg","relpath":"1_9_ab990f98","size":306926,"sizeH":["306.9","kB"],"date":"2019-08-05T12:24:13+00:00","type":"image/jpeg","extension":"jpg","params":{"width":1280,"height":960},"shortpath":"201908/1_9_ab990f98.jpg","width":1280,"height":960}]` {
		t.Fatal(string(d))
	}
}

func TestDBService_UrlByHash(t *testing.T) {
	ctx := context.Background()

	url, err := service.UrlByHash(ctx, "123456", "", "")
	if err != nil {
		t.Fatal(err)
	}

	if url != "1/23/123456.jpg" {
		t.Fatal(url)
	}
}

func TestDBService_UrlByHashList(t *testing.T) {
	ctx := context.Background()

	resp, err := service.UrlByHashList(ctx, []string{"123456", "987654"}, "", "")
	if err != nil {
		t.Fatal(err)
	}

	d, _ := json.Marshal(resp)
	if string(d) != `[{"hash":"123456","webPath":"1/23/123456.jpg"},{"hash":"987654","webPath":"9/87/987654.jpg"}]` {
		t.Fatal(string(d))
	}
}

func TestService_DeleteHash(t *testing.T) {
	ctx := context.Background()
	// hash upload
	data, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=")
	if err != nil {
		t.Errorf("failed to decode base64 image: %v", err)
	}
	path, err := testVfs.HashUpload(bytes.NewReader(data), testNs, "png")
	if err != nil {
		t.Errorf("failed to perform hash upload: %v", err)
	} else {
		t.Logf("hash=%v dir=%s file=%s", path, path.Dir(), path.File())
	}

	_, err = service.DeleteHash(ctx, "", path.Hash, path.Ext)
	if err.Error() != "Not Found" {
		t.Fatalf("deleting not existed hash err=%v", err)
	}

	_, err = service.DeleteHash(ctx, testNs, path.Hash, path.Ext)
	if err != nil {
		t.Fatalf("deleting existed hash err=%v", err)
	}
}
