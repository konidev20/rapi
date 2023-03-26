//go:build !windows
// +build !windows

package restorer

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/konidev20/restic-api/internal/repository"
	"github.com/konidev20/restic-api/internal/restic"
	rtest "github.com/konidev20/restic-api/internal/test"
)

func TestRestorerRestoreEmptyHardlinkedFileds(t *testing.T) {
	repo := repository.TestRepository(t)

	sn, _ := saveSnapshot(t, repo, Snapshot{
		Nodes: map[string]Node{
			"dirtest": Dir{
				Nodes: map[string]Node{
					"file1": File{Links: 2, Inode: 1},
					"file2": File{Links: 2, Inode: 1},
				},
			},
		},
	})

	res := NewRestorer(context.TODO(), repo, sn, false)

	res.SelectFilter = func(item string, dstpath string, node *restic.Node) (selectedForRestore bool, childMayBeSelected bool) {
		return true, true
	}

	tempdir := rtest.TempDir(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := res.RestoreTo(ctx, tempdir)
	rtest.OK(t, err)

	f1, err := os.Stat(filepath.Join(tempdir, "dirtest/file1"))
	rtest.OK(t, err)
	rtest.Equals(t, int64(0), f1.Size())
	s1, ok1 := f1.Sys().(*syscall.Stat_t)

	f2, err := os.Stat(filepath.Join(tempdir, "dirtest/file2"))
	rtest.OK(t, err)
	rtest.Equals(t, int64(0), f2.Size())
	s2, ok2 := f2.Sys().(*syscall.Stat_t)

	if ok1 && ok2 {
		rtest.Equals(t, s1.Ino, s2.Ino)
	}
}

func getBlockCount(t *testing.T, filename string) int64 {
	fi, err := os.Stat(filename)
	rtest.OK(t, err)
	st := fi.Sys().(*syscall.Stat_t)
	if st == nil {
		return -1
	}
	return st.Blocks
}