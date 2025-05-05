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
	domain      string
	serverPort  uint
	network     string
	adminTwinID uint64
}

// These variables are set during build time using ldflags
var (
	commit  string
	version string
)

func main() {
	if err := Run(); err != nil {
		log.Fatal().Err(err).Send()
	}
}

func Run() error {
	f := flags{}
	var sqlLogLevel int
	flag.StringVar(&f.Config.PostgresHost, "postgres-host", "", "postgres host")
	flag.Uint64Var(&f.Config.PostgresPort, "postgres-port", 5432, "postgres port")
	flag.StringVar(&f.Config.DBName, "postgres-db", "", "postgres database")
	flag.StringVar(&f.Config.PostgresUser, "postgres-user", "", "postgres username")
	flag.StringVar(&f.Config.PostgresPassword, "postgres-password", "", "postgres password")
	flag.StringVar(&f.Config.SSLMode, "ssl-mode", "disable", "postgres ssl mode[disable, require, verify-ca, verify-full]")
	flag.IntVar(&sqlLogLevel, "sql-log-level", 2, "sql logger level")
	flag.Uint64Var(&f.Config.MaxOpenConns, "max-open-conn", 3, "max open sql connections")
	flag.Uint64Var(&f.Config.MaxIdleConns, "max-idle-conn", 3, "max idle sql connections")

	flag.BoolVar(&f.version, "v", false, "shows the package version")
	flag.BoolVar(&f.debug, "debug", false, "allow debug logs")
	flag.UintVar(&f.serverPort, "server-port", 8080, "server port")
	flag.StringVar(&f.domain, "domain", "", "domain on which the server will be served")
	flag.StringVar(&f.network, "network", "dev", "the registrar network")
	flag.Uint64Var(&f.adminTwinID, "admin-twin-id", 1, "admin twin ID")

	flag.Parse()
	f.Config.SqlLogLevel = logger.LogLevel(sqlLogLevel)

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

	log.Info().Msgf("server is running on port :%d", f.serverPort)

	err = s.Run(fmt.Sprintf("%s:%d", f.domain, f.serverPort))
	if err != nil {
		return errors.Wrap(err, "failed to run gin server")
	}

	return nil
}

func (f flags) validate() error {
	if f.serverPort < 1 || f.serverPort > 65535 {
		return errors.Errorf("invalid port %d, server port should be in the valid port range 1–65535", f.serverPort)
	}

	if strings.TrimSpace(f.domain) == "" {
		return errors.New("invalid domain name, domain name should not be empty")
	}

	if f.Config.SqlLogLevel < 1 || f.Config.SqlLogLevel > 4 {
		return errors.Errorf("invalid sql log level %d, sql log level should be in the range 1-4", f.Config.SqlLogLevel)
	}
	if f.adminTwinID == 0 {
		return errors.Errorf("invalid admin twin id %d, admin twin id should not be 0", f.adminTwinID)
	}

	if _, err := net.LookupHost(f.domain); err != nil {
		return errors.Wrapf(err, "invalid domain %s", f.domain)
	}

	return f.Config.Validate()
}
