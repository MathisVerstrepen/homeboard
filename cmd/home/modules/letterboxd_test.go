package modules

import (
	"path/filepath"
	"runtime"
	"testing"

	f "github.com/MathisVerstrepen/go-module/webfetch"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func TestGetFriendsRecentMovies(t *testing.T) {
	type args struct {
		fetcher f.Fetcher
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test 1",
			args: args{
				fetcher: (*f.InitFetchers(filepath.Join(basepath, "/../../")))[0],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetFriendsRecentMovies(tt.args.fetcher)
		})
	}
}
