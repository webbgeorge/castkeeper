package e2e

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
)

func TestMain(m *testing.M) {
	var logBuf *bytes.Buffer

	flag.Parse() // must be called before you can use `testing.Short()` in TestMain
	if !testing.Short() {
		var killFn func()
		logBuf, killFn = startServer()
		defer killFn()
	}

	m.Run()

	if !testing.Short() && testing.Verbose() {
		fmt.Println("")
		log.Print("E2E test server logs: \n", logBuf.String(), "\n")
	}
}

func TestE2E(t *testing.T) {
	browser := rod.New().MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("http://localhost:8081").MustWaitStable()

	page.MustScreenshot("my.png")
}

func startServer() (logs *bytes.Buffer, killFn func()) {
	logBuf := bytes.Buffer{}
	_, filename, _, _ := runtime.Caller(0)
	configPath := filepath.Join(filepath.Dir(filename), "castkeeper.yml")
	cmdStr := filepath.Join(filepath.Dir(filename), "..", "cmd", "server", "server")

	buildCmd := exec.Command("make", "build")
	buildCmd.Dir = filepath.Join(filepath.Dir(filename), "..")
	if err := buildCmd.Start(); err != nil {
		log.Fatal("failed to start build server cmd in e2e tests:", err)
	}
	if err := buildCmd.Wait(); err != nil {
		log.Fatal("failed to build server in e2e tests:", err)
	}

	cmd := exec.Command(cmdStr, configPath)

	cmd.Stdout = &logBuf
	cmd.Stderr = &logBuf

	if err := cmd.Start(); err != nil {
		log.Fatal("failed to start server in e2e tests:", err)
	}

	// wait for server to have started
	i := 0
	for {
		if strings.Contains(logBuf.String(), "Starting server at ':8081'") {
			break
		}
		time.Sleep(time.Millisecond * 100)
		if i > 50 {
			fmt.Println("")
			log.Fatal("timed out waiting for http server to start: \n", logBuf.String(), "\n")
		}
		i++
	}

	return &logBuf, func() {
		log.Print("stopping server")
		if err := cmd.Process.Kill(); err != nil {
			log.Fatal("failed to kill server in e2e tests:", err)
		}
	}
}
