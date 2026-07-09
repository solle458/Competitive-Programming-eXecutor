package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var ErrConfigNotFound = errors.New("config file not found")

type File struct {
	RootDir        string   `yaml:"root_dir"`
	LibraryDirs    []string `yaml:"library_dirs"`
	DefaultLang    string   `yaml:"default_lang"`
	AtCoderSession string   `yaml:"atcoder_session,omitempty"`
}

type Config struct {
	File File
}

func NewConfig() *Config {
	return &Config{
		File: File{
			RootDir:        "",
			LibraryDirs:    []string{},
			DefaultLang:    "cpp",
			AtCoderSession: "",
		},
	}
}

func Load(root string) (*Config, error) {
	cfgPath := filepath.Join(root, ".cpx", "config.yaml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return nil, ErrConfigNotFound
	}
	cfgFile, err := os.Open(cfgPath)
	if err != nil {
		return nil, err
	}

	defer cfgFile.Close()
	decoder := yaml.NewDecoder(cfgFile)
	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Update(cfg *Config) error {
	cfgPath := filepath.Join(cfg.File.RootDir, ".cpx", "config.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0o755)
	cfgFile, err := os.Create(cfgPath)
	if err != nil {
		return err
	}
	defer cfgFile.Close()
	encoder := yaml.NewEncoder(cfgFile)
	if err := encoder.Encode(cfg); err != nil {
		return err
	}
	return encoder.Close()
}

func AddLibraryDir(cfg *Config, libraryDirs []string) error {
	newLibraryDirs := append(cfg.File.LibraryDirs, libraryDirs...)
	for _, dir := range newLibraryDirs {
		dir, err := filepath.Abs(dir)
		if err != nil {
			return err
		}
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return errors.New("library directory not found")
		}
	}
	cfg.File.LibraryDirs = newLibraryDirs
	return nil
}

func FindAndLoad() (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	for {
		cfg, err := Load(wd)
		if err != nil && !errors.Is(err, ErrConfigNotFound) {
			return nil, err
		}
		if cfg != nil {
			return cfg, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return nil, ErrConfigNotFound
		}
		wd = parent
	}
}
