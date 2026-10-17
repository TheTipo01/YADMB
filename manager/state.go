package manager

import (
	"github.com/TheTipo01/YADMB/database"
	"github.com/disgoorg/snowflake/v2"
)

// GetCustom returns a custom command by name.
func (server *Server) GetCustom(name string) (*database.CustomCommand, bool) {
	value, ok := server.Custom.Load(name)
	if !ok {
		return nil, false
	}

	command, ok := value.(*database.CustomCommand)
	return command, ok
}

// SetCustom stores a custom command.
func (server *Server) SetCustom(name string, command *database.CustomCommand) {
	server.Custom.Store(name, command)
}

// DeleteCustom removes a custom command.
func (server *Server) DeleteCustom(name string) {
	server.Custom.Delete(name)
}

// CustomNames returns all custom command names.
func (server *Server) CustomNames() []string {
	var names []string

	server.Custom.Range(func(key, _ any) bool {
		if name, ok := key.(string); ok {
			names = append(names, name)
		}
		return true
	})

	return names
}

// DJEnabled reports whether DJ mode is enabled.
func (server *Server) DJEnabled() bool {
	server.djMu.RLock()
	defer server.djMu.RUnlock()

	return server.DjMode
}

// SetDJEnabled enables or disables DJ mode.
func (server *Server) SetDJEnabled(enabled bool) {
	server.djMu.Lock()
	defer server.djMu.Unlock()

	server.DjMode = enabled
}

// DJRoleID returns the configured DJ role.
func (server *Server) DJRoleID() snowflake.ID {
	server.djMu.RLock()
	defer server.djMu.RUnlock()

	return server.DjRole
}

// SetDJRoleID sets the configured DJ role.
func (server *Server) SetDJRoleID(role snowflake.ID) {
	server.djMu.Lock()
	defer server.djMu.Unlock()

	server.DjRole = role
}
