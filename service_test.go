package vfs_test

import (
	"context"
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
)

func TestMain(m *testing.M) {
	flag.Parse()
	cfg, err := pg.ParseURL(*dbConn)
	if err != nil {
		panic(err)
	}

	dbc := pg.Connect(cfg)
	repo := db.NewVfsRepo(db.New(dbc))
	service = vfs.NewService(repo, vfs.VFS{}, dbc)
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
