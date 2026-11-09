package liqmap

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestEtherscanErrorHidesKey(t *testing.T) {
	s := &Service{http: &http.Client{Timeout: time.Second}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := s.borrowLogsOnce(ctx, "https://api.etherscan.io/v2/api?apikey=SECRETKEY")
	if err == nil {
		t.Fatal("expected an error from a cancelled request")
	}
	if strings.Contains(err.Error(), "SECRETKEY") {
		t.Fatalf("API key leaked into error: %q", err)
	}
}
