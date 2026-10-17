package manager

import (
	"sync"
	"testing"

	"github.com/disgoorg/snowflake/v2"
)

func TestServersGetOrCreateIsUnique(t *testing.T) {
	servers := NewServers()
	guildID := snowflake.ID(1)

	first := servers.GetOrCreate(guildID, nil)
	second := servers.GetOrCreate(guildID, nil)

	if first != second {
		t.Fatal("expected the same server instance")
	}

	if servers.Len() != 1 {
		t.Fatalf("expected one server, got %d", servers.Len())
	}

	servers.Remove(guildID)

	if servers.Len() != 0 {
		t.Fatalf("expected zero servers after Remove, got %d", servers.Len())
	}
}

func TestServersGetOrCreateConcurrent(t *testing.T) {
	servers := NewServers()
	guildID := snowflake.ID(1)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			servers.GetOrCreate(guildID, nil)
		}()
	}

	wg.Wait()

	if servers.Len() != 1 {
		t.Fatalf("expected one server, got %d", servers.Len())
	}
}
