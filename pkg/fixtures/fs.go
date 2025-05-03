package fixtures

import (
	"os"
	"path"

	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
)

func ConfigureFSForTestWithFixtures() (fs *os.Root, resetFn func()) {
	rootPath := path.Join(os.TempDir(), "castkeepertest", randomHex())
	root := objectstorage.MustOpenLocalFSRoot(rootPath)
	return root, func() {
		root.Close()
		os.RemoveAll(rootPath)
	}
	// TODO fixtures
	// 	"/testdataobj/xyz/xyz.jpg": fileFixture([]byte("CONTENTS")),
	// 	"/testdataobj/xyz/abc.mp3": fileFixture([]byte("CONTENTS")),
}
