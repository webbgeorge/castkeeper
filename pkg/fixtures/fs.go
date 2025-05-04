package fixtures

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
)

func ConfigureFSForTestWithFixtures() (fs *os.Root, resetFn func()) {
	rootPath := path.Join(os.TempDir(), "castkeepertest", randomHex())
	root := objectstorage.MustOpenLocalFSRoot(rootPath)

	// matching valid.xml fixture
	podImageFixtureFile(root, "916ed63b-7e5e-5541-af78-e214a0c14d95")
	podMP3FixtureFile(root,
		"916ed63b-7e5e-5541-af78-e214a0c14d95",
		"c8998fa5-8083-56a6-8d3c-7b98d031b3d8",
	)

	return root, func() {
		root.Close()
		os.RemoveAll(rootPath)
	}
}

func podImageFixtureFile(root *os.Root, podGUID string) {
	err := root.Mkdir(podGUID, 0750)
	if err != nil {
		if !errors.Is(err, fs.ErrExist) {
			panic(err)
		}
	}
	writeFixtureFile(
		root,
		fmt.Sprintf("%s/%s.jpg", podGUID, podGUID),
		[]byte("Not a real JPG"),
	)
}

func podMP3FixtureFile(root *os.Root, podGUID, epGUID string) {
	err := root.Mkdir(podGUID, 0750)
	if err != nil {
		if !errors.Is(err, fs.ErrExist) {
			panic(err)
		}
	}
	writeFixtureFile(
		root,
		fmt.Sprintf("%s/%s.mp3", podGUID, epGUID),
		[]byte("Not a real MP3"),
	)
}

func writeFixtureFile(root *os.Root, path string, content []byte) {
	f, err := root.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.Write(content)
	if err != nil {
		panic(err)
	}
}
