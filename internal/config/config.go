package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type LevelList []logrus.Level

func (a LevelList) MarshalText() ([]byte, error) {
	if len(a) == 0 {
		return []byte("-"), nil
	}

	var s string

	for i, e := range a {
		if i != 0 {
			s += ","
		}

		s += e.String()
	}

	return []byte(s), nil
}

func (a *LevelList) UnmarshalText(d []byte) error {
	if string(d) == "" || string(d) == "-" {
		*a = LevelList{}
		return nil
	}

	var aa LevelList

	for _, e := range strings.Split(string(d), ",") {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}

		l, err := logrus.ParseLevel(e)
		if err != nil {
			return fmt.Errorf("config.LevelList.UnmarshalText: could not parse value as logrus level: %w", err)
		}

		aa = append(aa, l)
	}

	*a = aa

	return nil
}

type LogQueries struct {
	Enabled    bool
	SlowerThan time.Duration
}

func (l LogQueries) String() string {
	if l.Enabled {
		if l.SlowerThan != 0 {
			return ">" + l.SlowerThan.String()
		}

		return "all"
	}

	return "none"
}

func (l LogQueries) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

func (l *LogQueries) UnmarshalText(d []byte) error {
	s := string(d)

	switch s {
	case "all":
		l.Enabled = true
		l.SlowerThan = 0
		return nil
	case "", "none":
		l.Enabled = false
		l.SlowerThan = 0
		return nil
	default:
		if s[0] == '>' && len(s) > 1 {
			d, err := time.ParseDuration(s[1:])
			if err != nil {
				return fmt.Errorf("config.LogQueries.UnmarshalText: could not parse value as duration: %w", err)
			}
			l.Enabled = true
			l.SlowerThan = d
			return nil
		}

		return fmt.Errorf("config.LogQueries.UnmarshalText: unrecognised input %q; valid options are none, all, or >x where x is a duration", s)
	}
}

func (l *LogQueries) IsZero() bool {
	return l.Enabled == false && l.SlowerThan == 0
}

type Config struct {
	Config               string       `name:"config" toml:"config" yaml:"config" help:"Config file location."`
	LogLevel             logrus.Level `name:"log_level" toml:"log_level" yaml:"log_level" help:"Global log level."`
	LogDebugLevels       LevelList    `name:"log_debug_levels" toml:"log_debug_levels" yaml:"log_debug_levels" help:"Which log levels to include stack data on."`
	LogQueries           LogQueries   `name:"log_queries" toml:"log_queries" yaml:"log_queries" help:"Log SQL queries."`
	LogSORM              bool         `name:"log_sorm" toml:"log_sorm" yaml:"log_sorm" help:"Log SORM queries."`
	ApplicationAddr      string       `name:"application_addr" toml:"application_addr" yaml:"application_addr" help:"Address to listen on for application server."`
	ApplicationDatabase  string       `name:"application_database" toml:"application_database" yaml:"application_database" help:"Database location for application."`
	ApplicationCachePath string       `name:"application_cache_path" toml:"application_cache_path" yaml:"application_cache_path" help:"Location for HTTP client cache."`
	ApplicationDataPath  string       `name:"application_data_path" toml:"application_data_path" yaml:"application_data_path" help:"Location for downloaded and converted media."`
	ApplicationMinify    bool         `name:"application_minify" toml:"application_minify" yaml:"application_minify" help:"Minify HTML/CSS/JS output."`
	BackgroundWorkers    int          `name:"background_workers" toml:"background_workers" yaml:"background_workers" help:"How many background workers to run."`
}

func (c Config) DataFile(section, name string) string {
	return filepath.Join(c.ApplicationDataPath, section, name)
}
