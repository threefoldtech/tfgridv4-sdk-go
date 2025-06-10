package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/db"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/server"
	"gorm.io/gorm/logger"
)

type flags struct {
	db.Config
	debug       bool
	version     bool
	host        string
	port        uint
	network     string
	adminTwinID uint64
}

// These variables are set during build time using ldflags
var (
	commit  string
	version string
)

// @title Node Registrar API
// @version 1.0
// @description API for managing TFGrid node registration
// @BasePath /api/v1

func main() {
	if err := Run(); err != nil {
		log.Fatal().Err(err).Send()
	}
}

func Run() error {
	f := flags{}
	var sqlLogLevel int
	flag.StringVar(&f.PostgresHost, "postgres-host", "", "postgres host")
	flag.Uint64Var(&f.PostgresPort, "postgres-port", 5432, "postgres port")
	flag.StringVar(&f.DBName, "postgres-db", "", "postgres database")
	flag.StringVar(&f.PostgresUser, "postgres-user", "", "postgres username")
	flag.StringVar(&f.PostgresPassword, "postgres-password", "", "postgres password")
	flag.StringVar(&f.SSLMode, "ssl-mode", "disable", "postgres ssl mode[disable, require, verify-ca, verify-full]")
	flag.IntVar(&sqlLogLevel, "sql-log-level", 2, "sql logger level")
	flag.Uint64Var(&f.MaxOpenConns, "max-open-conn", 3, "max open sql connections")
	flag.Uint64Var(&f.MaxIdleConns, "max-idle-conn", 3, "max idle sql connections")

	flag.BoolVar(&f.version, "v", false, "shows the package version")
	flag.BoolVar(&f.debug, "debug", false, "allow debug logs")
	flag.UintVar(&f.port, "server-port", 8080, "server port")
	flag.StringVar(&f.host, "host", "", "host on which the server will be served")

	// Deprecated flag handling
	flag.Func("domain", "deprecated: use --host instead", func(val string) error {
		log.Warn().Msg("Warning: --domain flag is deprecated, please use --host instead")
		f.host = val
		return nil
	})

	flag.StringVar(&f.network, "network", "dev", "the registrar network")
	flag.Uint64Var(&f.adminTwinID, "admin-twin-id", 1, "admin twin ID")

	flag.Parse()
	f.SqlLogLevel = logger.LogLevel(sqlLogLevel)

	if f.version {
		log.Info().Str("version", version).Str("commit", commit).Send()
		return nil
	}

	if err := f.validate(); err != nil {
		return err
	}

	logLevel := zerolog.InfoLevel
	if f.debug {
		logLevel = zerolog.DebugLevel
	}
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Level(logLevel).With().Timestamp().Logger()

	db, err := db.NewDB(f.Config)
	if err != nil {
		return errors.Wrap(err, "failed to open database with the specified configurations")
	}

	defer func() {
		err = db.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close database connection")
		}
	}()

	s := server.NewServer(db, f.network, f.adminTwinID)

	log.Info().Msgf("server is running on port :%d", f.port)

	err = s.Run(fmt.Sprintf("%s:%d", f.host, f.port))
	if err != nil {
		return errors.Wrap(err, "failed to run gin server")
	}

	return nil
}

func (f flags) validate() error {
	if f.port < 1 || f.port > 65535 {
		return errors.Errorf("invalid port %d, server port should be in the valid port range 1–65535", f.port)
	}

	if f.SqlLogLevel < 1 || f.SqlLogLevel > 4 {
		return errors.Errorf("invalid sql log level %d, sql log level should be in the range 1-4", f.SqlLogLevel)
	}

	if f.adminTwinID == 0 {
		return errors.Errorf("invalid admin twin id %d, admin twin id should not be 0", f.adminTwinID)
	}

	if err := f.validateHost(); err != nil {
		return err
	}

	return f.Validate()
}

func (f flags) validateHost() error {
	host := strings.TrimSpace(f.host)
	if host == "" {
		return errors.New("host cannot be empty")
	}

	// Check common binding addresses
	switch host {
	case "localhost", "0.0.0.0", "127.0.0.1":
		return nil
	}

	// Check if valid IP address
	if ip := net.ParseIP(host); ip != nil {
		return nil
	}

	// Check if valid hostname
	if _, err := net.LookupHost(host); err != nil {
		return errors.Wrapf(err, "invalid host %q: must be a valid IP address, hostname, or domain name", host)
	}

	return nil
}
