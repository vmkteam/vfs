package vfs

import (
	"testing"
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
