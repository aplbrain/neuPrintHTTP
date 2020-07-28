package config

// loads all storage plugins
import (
	"github.com/aplbrain/neuPrintHTTP/storage"
	_ "github.com/aplbrain/neuPrintHTTP/storage/badger"
	_ "github.com/aplbrain/neuPrintHTTP/storage/dvid"
	_ "github.com/aplbrain/neuPrintHTTP/storage/dvidkv"
	_ "github.com/aplbrain/neuPrintHTTP/storage/neuprintneo4j"
)

// CreateStore creates a datastore from the engine specified by the configuration
func CreateStore(config Config) (storage.Store, error) {
	if config.Timeout == 0 {
		config.Timeout = 60
	}
	return storage.ParseConfig(config.Engine, config.EngineConfig, config.MainStores, config.DataTypes, config.Timeout)
}
