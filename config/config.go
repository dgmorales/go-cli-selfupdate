package config

// ServerSideConfig represents configuration we store on the server side of our CLI application
//
// These are things that do not make sense to store in a local config, like the minimal
// required version this app needs to be in, to be able to talk to our current API
// servers.
type ServerSideConfig struct {
	RepoOwner              string
	RepoName               string
	MinimalRequiredVersion string
}

// ServerSideConfigLoader knows how to reach, read and parse our server side config.
type ServerSideConfigLoader interface {
	Load() (ServerSideConfig, error)
}
