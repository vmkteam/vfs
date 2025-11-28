package vfssrv

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vmkteam/appkit"
)

func TestNewClient_Example(t *testing.T) {
	t.SkipNow()

	c := NewClient(Opts{
		ApiURL:    "http://localhost:9999/",
		PublicURL: "http://localhost:9999/media/",
		Client:    appkit.NewHTTPClient("appsrv", "v1", time.Second*20),
	})

	ctx := t.Context()
	ctx = context.WithValue(ctx, echo.HeaderXRequestID, "testreq1") //nolint:staticcheck

	// get token
	token, err := c.AuthToken(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token, err)

	// send file
	f, err := os.Open("../../image.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	r, err := c.UploadFile(ctx, token, "", "test.gif", f)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(r)

	// download file
	img, er := c.DownloadImage(ctx, "", r.Hash, "")
	t.Log(len(img), er)
}

func TestClient_FilePath(t *testing.T) {
	type args struct {
		namespace string
		hash      string
		size      string
		ext       string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "simple example",
			args: args{
				namespace: "",
				hash:      "64a9f060983200709061894cc5f69f83",
				size:      "",
				ext:       "",
			},
			want:    "http://localhost:9999/media/6/4a/64a9f060983200709061894cc5f69f83.jpg",
			wantErr: false,
		},
		{
			name: "full example",
			args: args{
				namespace: "items",
				hash:      "64a9f060983200709061894cc5f69f83",
				size:      "full",
				ext:       "pdf",
			},
			want:    "http://localhost:9999/media/items/full/6/4a/64a9f060983200709061894cc5f69f83.pdf",
			wantErr: false,
		},
	}

	mediaURL := "http://localhost:9999/media/"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(Opts{ApiURL: mediaURL})
			got, err := c.FilePath(tt.args.namespace, tt.args.hash, tt.args.size, tt.args.ext)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FilePath() got = %v, want %v", got, tt.want)
			}
		})
	}
}
