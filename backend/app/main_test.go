package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		goleak.IgnoreTopFunction("github.com/theshamuel/data-sec-exchanger/backend/app.init.0.func1"),
		goleak.IgnoreTopFunction("net/http.(*Server).Shutdown"))
}

func Test_Main(t *testing.T) {
	port := 3000
	fmt.Println(port)
	done := make(chan struct{})
	go func() {
		<-done
		e := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		require.NoError(t, e)
	}()

	finished := make(chan struct{})

	go func() {
		main()
		close(finished)
	}()

	defer func() {
		close(done)
		<-finished
	}()
	os.Args = []string{"--debug"}
	waitForHTTPServer(port)
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ping", port))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "pong\n", string(body))
}

func TestGetStackTrace(t *testing.T) {
	stackTrace := getStackTrace()
	assert.True(t, strings.Contains(stackTrace, "goroutine"))
	assert.True(t, strings.Contains(stackTrace, "[running]"))
	assert.True(t, strings.Contains(stackTrace, "data-sec-exchanger/backend/app/main.go"))
	assert.True(t, strings.Contains(stackTrace, "data-sec-exchanger/backend/app.getStackTrace"))
	t.Logf("\n STACKTRACE: %s", stackTrace)
}

func generateRndPort() (port int) {
	for {
		rand.Seed(time.Now().UnixNano())
		port = 49001 + int(rand.Int31n(150))
		if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err == nil {
			_ = ln.Close()
			break
		}
	}
	return port
}

func waitForHTTPServer(port int) {
	client := http.Client{Timeout: time.Second}
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second * 5)
		if resp, err := client.Get(fmt.Sprintf("http://localhost:%d/ping", port)); err == nil {
			_ = resp.Body.Close()
			return
		}
	}
}
