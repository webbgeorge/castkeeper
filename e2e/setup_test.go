package e2e

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func setupE2ETests(verbose, debug bool) (browser *rod.Browser, cleanup func()) {
	deleteDatabase()
	err := createTestUser()
	if err != nil {
		panic(err)
	}

	logBuf, killServer, err := startServer()
	if err != nil {
		panic(err)
	}

	browser, cleanupBrowser := setupBrowser(debug)

	return browser, func() {
		cleanupBrowser()
		killServer()
		deleteDatabase()
		if verbose {
			fmt.Println("")
			log.Print("E2E test server logs: \n", logBuf.String(), "\n")
		}
	}
}

func startServer() (logBuf *bytes.Buffer, killFn func(), err error) {
	cmd, logBuf := createGoRunCmd("server")
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start server in e2e tests '%w'", err)
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
			return nil, nil, fmt.Errorf("timed out waiting for http server to start: \n %s \n", logBuf.String())
		}
		i++
	}

	return logBuf, func() {
		log.Print("stopping server")
		if err := cmd.Process.Kill(); err != nil {
			log.Print("failed to kill server in e2e tests:", err)
		}
	}, nil
}

func createGoRunCmd(cmdName string, arg ...string) (*exec.Cmd, *bytes.Buffer) {
	_, filename, _, _ := runtime.Caller(0)
	configPath := filepath.Join(filepath.Dir(filename), "castkeeper.yml")
	cmdStr := filepath.Join(filepath.Dir(filename), "..", "cmd", cmdName)

	cmdArg := []string{"run", cmdStr, configPath}
	cmdArg = append(cmdArg, arg...)

	cmd := exec.Command("go", cmdArg...)

	logBuf := &bytes.Buffer{}
	cmd.Stdout = logBuf
	cmd.Stderr = logBuf

	return cmd, logBuf
}

func createTestUser() error {
	username := "e2euser"
	password := "e2epass"

	cmd, logBuf := createGoRunCmd("createuser", username, password)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create user in e2e tests '%s', cmd logs: %s", err.Error(), logBuf.String())
	}
	return nil
}

func deleteDatabase() {
	_, filename, _, _ := runtime.Caller(0)
	dbPath := filepath.Join(filepath.Dir(filename), "..", "data", "test-e2e.db")
	_ = os.Remove(dbPath)
}

func setupBrowser(debug bool) (*rod.Browser, func()) {
	if !debug {
		browser := rod.New().MustConnect()
		return browser, func() {
			browser.MustClose()
		}
	}

	l := launcher.New().
		Headless(false).
		Devtools(true)

	url := l.MustLaunch()

	browser := rod.New().
		ControlURL(url).
		Trace(true).
		SlowMotion(time.Second).
		NoDefaultDevice().
		MustConnect()

	cleanup := func() {
		l.Cleanup()
		browser.MustClose()
	}

	return browser, cleanup
}
