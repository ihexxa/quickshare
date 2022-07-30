package fileindex

import (
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ihexxa/quickshare/src/fs/local"
	"github.com/ihexxa/quickshare/src/idgen/simpleidgen"
	"github.com/ihexxa/randstr"
)

func TestFileSearch(t *testing.T) {
	dirPath := "tmp"
	err := os.MkdirAll(dirPath, 0700)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirPath)

	makePaths := func(maxPathLen, count int) map[string]bool {
		rand.Seed(time.Now().UnixNano())
		randStr := randstr.New([]string{})

		paths := map[string]bool{}
		for i := 0; i < count; i++ {
			pathLen := rand.Intn(maxPathLen) + 1
			pathParts := []string{}
			for j := 0; j < pathLen; j++ {
				pathParts = append(pathParts, randStr.Alnums())
			}
			paths[strings.Join(pathParts, "/")] = true
		}

		return paths
	}

	ider := simpleidgen.New()
	fs := local.NewLocalFS(dirPath, 0660, 1024, 60, 60, ider)
	fileIndex := NewFileTreeIndex(fs, "/", 0)

	paths := makePaths(8, 256)
	for pathname := range paths {
		err := fileIndex.AddPath(pathname)
		if err != nil {
			t.Fatal(err)
		}
	}

	indexPath := "/fileindex"
	err = fileIndex.WriteTo(indexPath)
	if err != nil {
		t.Fatal(err)
	}

	fileIndex2 := NewFileTreeIndex(fs, "/", 0)
	err = fileIndex2.ReadFrom(indexPath)
	if err != nil {
		t.Fatal(err)
	}
}
