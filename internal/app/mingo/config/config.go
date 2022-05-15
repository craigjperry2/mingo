package config

import (
	"io"
	"strconv"
	"time"

	"github.com/craigjperry2/mingo/internal/app/mingo/database"
	"github.com/craigjperry2/mingo/internal/app/mingo/system"
)

type Config struct {
	progname           string
	startUtc           time.Time
	args               []string
	username           string
	hostname           string
	listenPort         uint16
	loggingDestination io.Writer
	staticDir          string
	clock              system.Clock
	db                 *database.DbNothingBurger
}

var instance *Config

func defaults() *Config {
	return &Config{
		progname:   "mingo",
		startUtc:   system.NewClock()().UTC(),
		listenPort: 8080,
		clock:      system.NewClock(),
	}
}

func Build(args []string, stderr io.Writer) (err error) {
	if instance != nil {
		panic("programmer error: attempt to re-build config")
	}
	cfg := defaults()

	cfg.args = args

	username, err := system.Username()
	if err != nil {
		return err
	}
	cfg.username = username

	hostname, err := system.Hostname()
	if err != nil {
		return err
	}
	cfg.hostname = hostname

	cfg.loggingDestination = stderr

	cfg.db = database.NewDatabase()

	instance, err = parseFlags(cfg)

	return err
}

func GetInstance() *Config {
	return instance
}

func (c *Config) GetHostname() string {
	return c.hostname
}

func (c *Config) GetLoggingDestination() io.Writer {
	return c.loggingDestination
}

func (c *Config) GetClock() system.Clock {
	return c.clock
}

func (c *Config) GetStartUtc() time.Time {
	return c.startUtc
}

func (c *Config) GetStaticDir() string {
	return c.staticDir
}

func (c *Config) GetListenPort() uint16 {
	return c.listenPort
}

func (c *Config) GetListenPortStr() string {
	return strconv.Itoa(int(c.listenPort))
}

func (c *Config) GetProgname() string {
	return c.progname
}

func (c *Config) GetUsername() string {
	return c.username
}

func (c *Config) GetDatabase() *database.DbNothingBurger {
	return c.db
}
