package config_test

import (
	"testing"

	"github.com/dgmorales/go-cli-selfupdate/config"
)

func TestKubeServerSideConfigLoaderImplementsServerSideConfigLoader(t *testing.T) {
	var i interface{} = new(config.KubeServerSideConfigLoader)
	if _, ok := i.(config.ServerSideConfigLoader); !ok {
		t.Fatalf("expected %T to implement ServerSideConfigLoader", i)
	}
}
