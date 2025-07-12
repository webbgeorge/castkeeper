package e2e

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

const (
	configProfileLocal = "local"
	configProfileS3    = "s3"
	e2eUsername        = "e2euser"
	e2ePassword        = "stronge2epass"
)

func setupE2ETests(configProfile string, verbose, debug bool) (browser *rod.Browser, cleanup func()) {
	deleteDatabase(configProfile)
	err := createTestUser(configProfile)
	if err != nil {
		panic(err)
	}

	logBuf, killServer, err := startServer(configProfile)
	if err != nil {
		panic(err)
	}

	browser, cleanupBrowser := setupBrowser(debug)

	return browser, func() {
		cleanupBrowser()
		killServer()
		deleteDatabase(configProfile)
		if verbose {
			fmt.Println("")
			log.Print("E2E test server logs: \n", logBuf.String(), "\n")
		}
	}
}

func startServer(configProfile string) (logBuf *bytes.Buffer, killFn func(), err error) {
	cmd, logBuf := createGoRunCmd(configProfile, "serve")
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start server in e2e tests '%w'", err)
	}

	// wait for server to have started
	i := 0
	for !strings.Contains(logBuf.String(), "Server listening at") {
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

func createGoRunCmd(configProfile, cmdName string, arg ...string) (*exec.Cmd, *bytes.Buffer) {
	_, filename, _, _ := runtime.Caller(0)
	configPath := filepath.Join(
		filepath.Dir(filename),
		fmt.Sprintf("castkeeper.%s.yml", configProfile),
	)
	cmdStr := filepath.Join(filepath.Dir(filename), "..", "cmd", "main.go")

	cmdArg := []string{"run", cmdStr, cmdName, "--config", configPath}
	cmdArg = append(cmdArg, arg...)

	cmd := exec.Command("go", cmdArg...)

	logBuf := &bytes.Buffer{}
	cmd.Stdout = logBuf
	cmd.Stderr = logBuf

	return cmd, logBuf
}

func createTestUser(configProfile string) error {
	cmd, logBuf := createGoRunCmd(
		configProfile,
		"user",
		"create",
		"--username", e2eUsername,
		"--password", e2ePassword,
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create user in e2e tests '%s', cmd logs: %s", err.Error(), logBuf.String())
	}
	return nil
}

func deleteDatabase(configProfile string) {
	_, filename, _, _ := runtime.Caller(0)
	dataDir := ""
	switch configProfile {
	case configProfileLocal:
		dataDir = "e2e-local"
	case configProfileS3:
		dataDir = "e2e-s3"
	}
	dbPath := filepath.Join(filepath.Dir(filename), "..", "data", dataDir, "data.db")
	_ = os.Remove(dbPath)
}

func setupBrowser(debug bool) (*rod.Browser, func()) {
	isCI, _ := strconv.ParseBool(os.Getenv("CI"))
	l := launcher.New().
		NoSandbox(isCI)

	if debug {
		l = l.
			Headless(false).
			Devtools(true)
	}

	url := l.MustLaunch()

	browser := rod.New().
		ControlURL(url)

	if debug {
		browser = browser.
			Trace(true).
			SlowMotion(time.Second).
			NoDefaultDevice()
	}

	browser.MustConnect()

	cleanup := func() {
		browser.MustClose()
		l.Cleanup()
	}

	return browser, cleanup
}
