package test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	cronet "github.com/sagernet/cronet-go"
	"github.com/stretchr/testify/require"
)

// Regression for chromium crbug.com/572055997 using valid chunked HTTP framing.
// The cache must not deliver bytes beyond its Content-Range, with or without storage.
func TestRangeResponseBoundaries(t *testing.T) {
	path := os.Getenv("NHOP_CRONET_LIBRARY")
	if path == "" {
		t.Skip("requires an explicitly selected native library")
	}
	require.NoError(t, cronet.LoadLibrary(path))
	for _, noStore := range []bool{false, true} {
		for _, split := range []bool{false, true} {
			t.Run(fmt.Sprintf("no-store=%t/split=%t", noStore, split), func(t *testing.T) {
				transport := rangeTransport(t)
				readRange(t, transport, noStore, split, true)
				readRange(t, transport, noStore, split, false)
			})
		}
	}
}

func rangeTransport(t *testing.T) *cronet.RoundTripper {
	t.Helper()
	params := cronet.NewEngineParams()
	defer params.Destroy()
	params.SetHTTPCacheMode(cronet.HTTPCacheModeInMemory)
	engine := cronet.NewEngine()
	require.Equal(t, cronet.ResultSuccess, engine.StartWithParams(params))
	executor := cronet.NewExecutor(func(_ cronet.Executor, command cronet.Runnable) {
		go func() {
			defer command.Destroy()
			command.Run()
		}()
	})
	t.Cleanup(func() {
		engine.Shutdown()
		engine.Destroy()
		executor.Destroy()
	})
	return &cronet.RoundTripper{Engine: engine, Executor: executor}
}

func readRange(t *testing.T, transport *cronet.RoundTripper, noStore, split, extra bool) {
	t.Helper()
	server := rangeServer(noStore, extra)
	defer server.Close()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	require.NoError(t, err)
	request.Header.Set("Range", "bytes=0-9")
	response, err := transport.RoundTrip(request)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusPartialContent, response.StatusCode)
	var first []byte
	if split {
		first = make([]byte, 10)
		_, err = io.ReadFull(response.Body, first)
		require.NoError(t, err)
		require.Equal(t, "0123456789", string(first))
	}
	rest, err := io.ReadAll(response.Body)
	checkRangeRead(t, first, rest, err, split, extra)
}

func checkRangeRead(t *testing.T, first, rest []byte, err error, split, extra bool) {
	t.Helper()
	if extra {
		// URLRequestHttpJob clears this error after exactly the advertised range
		// was delivered; the public API must still expose ZERO excess bytes.
		if len(rest) != 0 || (!split && (err == nil || !strings.Contains(err.Error(), "CONTENT_LENGTH_MISMATCH"))) ||
			(split && err != nil && !strings.Contains(err.Error(), "CONTENT_LENGTH_MISMATCH")) {
			t.Errorf("overlong range accepted: first=%q rest=%q error=%v", first, rest, err)
		}
	} else {
		require.NoError(t, err)
		require.Equal(t, "0123456789", string(append(first, rest...)))
	}
}

func rangeServer(noStore, extra bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body := "0123456789"
		if extra {
			body += "extra-data"
		}
		writer.Header().Set("Content-Range", "bytes 0-9/80")
		// Stream chunks so HTTP framing does not preempt the cache range check.
		// A contradictory Content-Length makes PartialData bypass the tested path.
		writer.Header().Set("ETag", "\"range-fixture\"")
		writer.Header().Set("Cache-Control", "max-age=600")
		if noStore {
			writer.Header().Set("Cache-Control", "no-store")
		}
		writer.WriteHeader(http.StatusPartialContent)
		writer.(http.Flusher).Flush()
		_, _ = io.WriteString(writer, body)
	}))
}
