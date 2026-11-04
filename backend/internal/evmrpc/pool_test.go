package evmrpc

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
)

func TestEndpointsAppendsBackupsWithoutDuplicates(t *testing.T) {
	got := Endpoints(" https://my-node.example/ , https://ethereum.publicnode.com", "ethereum")
	want := []string{"https://my-node.example", "https://ethereum.publicnode.com", "https://eth.drpc.org", "https://1rpc.io/eth"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if len(Endpoints("", "bsc")) != len(Backups["bsc"]) {
		t.Fatal("empty config should fall back to backups")
	}
}

func TestDoFailsOverAndRemembersWorkingNode(t *testing.T) {
	p := NewPool()
	defer p.Close()
	urls := []string{"http://127.0.0.1:1", "http://127.0.0.1:2", "http://127.0.0.1:3"}
	clients := map[*ethclient.Client]string{}
	for _, u := range urls {
		c, err := p.Client(context.Background(), u)
		if err != nil {
			t.Fatal(err)
		}
		clients[c] = u
	}

	var tried []string
	healthy := "http://127.0.0.1:2"
	fn := func(c *ethclient.Client) error {
		tried = append(tried, clients[c])
		if clients[c] != healthy {
			return errors.New("down")
		}
		return nil
	}
	if err := p.Do(context.Background(), urls, fn); err != nil {
		t.Fatal(err)
	}
	if want := urls[:2]; !reflect.DeepEqual(tried, want) {
		t.Fatalf("first round tried %v, want %v", tried, want)
	}

	tried = nil
	if err := p.Do(context.Background(), urls, fn); err != nil {
		t.Fatal(err)
	}
	if want := []string{healthy}; !reflect.DeepEqual(tried, want) {
		t.Fatalf("second round tried %v, want %v (should start from the node that worked)", tried, want)
	}

	healthy = "none"
	if err := p.Do(context.Background(), urls, fn); err == nil {
		t.Fatal("expected error when every node is down")
	}
}
