package manager

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/disgoorg/snowflake/v2"
)

// Servers is a concurrent registry of guild servers.
type Servers struct {
	m     sync.Map
	count atomic.Int64
}

func NewServers() *Servers {
	return &Servers{}
}

// Get returns the server for the given guild, if it exists.
func (s *Servers) Get(guildID snowflake.ID) (*Server, bool) {
	value, ok := s.m.Load(guildID)
	if !ok {
		return nil, false
	}

	return value.(*Server), true
}

// GetOrCreate returns the server for the given guild, creating it if needed.
func (s *Servers) GetOrCreate(guildID snowflake.ID, clients *Clients) *Server {
	if value, ok := s.m.Load(guildID); ok {
		return value.(*Server)
	}

	server := NewServer(guildID, clients)
	actual, loaded := s.m.LoadOrStore(guildID, server)
	if !loaded {
		s.count.Add(1)
		return server
	}

	return actual.(*Server)
}

// Remove removes the server for the given guild.
func (s *Servers) Remove(guildID snowflake.ID) {
	if _, loaded := s.m.LoadAndDelete(guildID); loaded {
		s.count.Add(-1)
	}
}

// Len returns the number of registered guilds.
func (s *Servers) Len() int {
	return int(s.count.Load())
}

// Range calls f for every registered server.
func (s *Servers) Range(f func(guildID snowflake.ID, server *Server) bool) {
	s.m.Range(func(key, value any) bool {
		return f(key.(snowflake.ID), value.(*Server))
	})
}

// Shutdown stops every registered server in parallel, waiting until done or
// until the context expires.
func (s *Servers) Shutdown(ctx context.Context) {
	var wg sync.WaitGroup

	s.Range(func(_ snowflake.ID, server *Server) bool {
		wg.Add(1)
		go func() {
			defer wg.Done()
			server.Shutdown(ctx)
		}()
		return true
	})

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
}
