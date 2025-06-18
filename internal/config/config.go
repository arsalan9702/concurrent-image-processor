package config

import (
	"errors"
	"runtime"

	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	InputDir    string  `mapstructure:"input_dir"`
	OutputDir   string  `mapstructure:"output_dir"`
	Filter      string  `mapstructure:"filter"`
	Workers     int     `mapstructure:"workers"`
	RowWorkers  int     `mapstructure:"row_workers"`
	Quality     int     `mapstructure:"quality"`
	BlurRadius  float64 `mapstructure:"blur_radius"`
	Brightness  float64 `mapstructure:"brightness"`
	Contrast    float64 `mapstructure:"contrast"`
	MaxFileSize int64   `mapstructure:"max_file_size"`
	BufferSize  int     `mapstructure:"buffer_size"`
}

// Load loads configuration from file and sets defaults
func Load(configFile string) (*Config, error) {
	// defaults
	viper.SetDefault("input_dir", "examples/images")
	viper.SetDefault("output_dir", "examples/output")
	viper.SetDefault("filter", "grayscale")
	viper.SetDefault("workers", runtime.NumCPU())
	viper.SetDefault("row_workers", runtime.NumCPU()*2)
	viper.SetDefault("quality", 95)
	viper.SetDefault("blur_radius", 2.0)
	viper.SetDefault("brightness", 1.2)
	viper.SetDefault("contrast", 1.1)
	viper.SetDefault("max_file_size", 100*1024*1024)
	viper.SetDefault("buffer_size", 1000)

	// Load config
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	// environment variable support
	viper.SetEnvPrefix("IMG_PROC")
	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// func to valuidate the configuration
func (c *Config) Validate() error {
	if c.Workers <= 0 {
		return errors.New("workers must be greater than 0")
	}
	if c.RowWorkers<=0{
		return errors.New("row_workers must be greater than 0")
	}
	if c.Quality<0 || c.Quality>100{
		return errors.New("quality must be between 1 and 100")
	}
	if c.BlurRadius<0{
		return errors.New("blur_radius must be non-zero")
	}
	if c.Brightness<=0{
		return errors.New("brightness must be greater than 0")
	}
	if c.MaxFileSize<=0{
		return errors.New("max_file_size must be greater than 0")
	}
	if c.BufferSize<=0{
		return errors.New("buffer_size must be greater than 0")
	}

	validFilters := map[string]bool{
		"grayscale": true,
		"blur": true,
		"brightness": true,
		"contrast": true,
	}
	if !validFilters[c.Filter]{
		return errors.New("invalid filter: must be grayscale, blur, brightness, or contrast")
	}

	return nil
}
