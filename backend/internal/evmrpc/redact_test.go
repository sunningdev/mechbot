package evmrpc

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
)

func TestRedactDropsKey(t *testing.T) {
	err := redact(&url.Error{Op: "Post", URL: "https://eth-mainnet.example.com/v2/SECRETKEY?key=SECRET2", Err: context.DeadlineExceeded})
	if msg := err.Error(); strings.Contains(msg, "SECRET") || !strings.Contains(msg, "eth-mainnet.example.com") {
		t.Fatalf("unexpected message: %q", msg)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("underlying error should stay wrapped")
	}
	plain := errors.New("execution reverted")
	if redact(plain) != plain {
		t.Fatal("non-URL errors must pass through unchanged")
	}
}
