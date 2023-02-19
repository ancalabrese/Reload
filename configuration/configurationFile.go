package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ConfigurationFile struct {
	filePath string
	config   interface{}
}

func NewConfigurationFile(
	path string,
	configuration interface{}) (*ConfigurationFile, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// At this point only files are supported. Make sure the path is not a folder
	if isDirectory(f) {
		return nil, fmt.Errorf("%s is a directory or not supported file type", path)
	}

	return &ConfigurationFile{
		filePath: path,
		config:   configuration,
	}, nil
}

func (cf *ConfigurationFile) LoadConfiguration() error {
	c, err := os.Open(cf.filePath)
	if err != nil {
		return fmt.Errorf("failed to open config file %s: %w", cf.filePath, err)
	}
	defer c.Close()

	err = json.NewDecoder(c).Decode(cf.config)
	if err != nil {
		return fmt.Errorf("[loadConfiguration] - failed to marshal new config: %w", err)
	}

	return nil
}

// isDirectory checks whether the provided files is a directory.
// Directories are not supported.
func isDirectory(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return true //error occurred, assuming this is not a supported file
	}

	return stat.IsDir()

}
