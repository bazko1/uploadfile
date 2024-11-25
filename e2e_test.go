package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// an e2e test for POST /api/v1/file and GET /api/v1/file/download/id
func TestUploadFileUsingAPI(t *testing.T) {
	fileJSON := ` {
    "name": "file.txt",
    "uploaded_by": "File Owner",
    "email": "test@test.com",
    "data": "aGVsbG8gd29ybGQK"
  } `

	dbFile := "test_upload_file.db"
	defer os.Remove(dbFile)
	e := echo.New()
	defer e.Close()
	createFileRouting(e, dbFile)
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create new listener error: %v", err)
	}
	defer ln.Close()
	e.Listener = ln
	serverServeError := make(chan error)

	go func() {
		if err := e.Start(""); err != nil {
			serverServeError <- err
		}
	}()

	// one time check if errors occurred or continue
	select {
	case err := <-serverServeError:
		t.Fatalf("Failed to start server error %v:", err)
	default:
		break
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := e.ListenerAddr().String()

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		fmt.Sprintf("http://%s/api/v1/file", addr),
		strings.NewReader(fileJSON))
	assert.NoError(t, err)

	req.SetBasicAuth("foo", "bar")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error during post call user recreation: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err = http.NewRequestWithContext(ctx,
		http.MethodGet,
		fmt.Sprintf("http://%s/api/v1/file/download/1", addr), nil)
	assert.NoError(t, err)
	req.SetBasicAuth("foo", "bar")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error during post call user recreation: %v", err)
	}
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "hello world\n", string(bytes))
}
