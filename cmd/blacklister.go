package main

import (
	"flag"
	"os"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gp42/pf-demo/pkg/api"
	"github.com/gp42/pf-demo/pkg/db"
	m "github.com/gp42/pf-demo/pkg/mailer"
	"github.com/gp42/pf-demo/pkg/util"
)

func main() {
	dbConnectionParams := db.DBConnectionParams{}
	mailer := &m.Mailer{}

	flagLogLevel := zap.LevelFlag("log-level", zapcore.InfoLevel, "Set log level. The following levels are available: 'debug', 'info', 'warn', 'error', 'panic' and 'fatal'")
	flagDevLog := flag.Bool("dev", false, "Enable development logger mode")
	flagListenAddress := flag.String("listen-address", "0.0.0.0:8080", "Interface and port to listen for connections")
	flagCertPath := flag.String("cert-path", "", "Certificate path for server TLS")
	flagCertKeyPath := flag.String("cert-key-path", "", "Certificate key path for server TLS")

	dbConnectionParams.Host = flag.String("db-host", "localhost", "Database connection host")
	dbConnectionParams.Port = flag.Int("db-port", 5432, "Database connection port")
	dbConnectionParams.DBName = flag.String("db-name", "blacklister", "Database name")
	dbConnectionParams.User = flag.String("db-user", "postgres", "Database user")
	dbConnectionParams.Password = flag.String("db-password", "postgres", "Database password")
	dbConnectionParams.SSLMode = flag.String("db-sslmode", "require", "Database connection sslmode. Partially implemented, only 'disable' and 'require' are supported.")

	mailer.From = flag.String("mail-from", "test@domain.com", "Send emails FROM email")
	var mailerToFlag util.ArrayFlag
	flag.Var(&mailerToFlag, "mail-to", "Send emails TO email")
	mailer.To = []string(mailerToFlag)
	mailer.Password = flag.String("mail-password", "", "Send emails password")
	mailer.SMTPHost = flag.String("mail-smtphost", "", "Send emails SMTP host")
	mailer.SMTPPort = flag.Int("mail-smtpport", 587, "Send emails SMTP port")
	flag.Parse()

	util.FlagsFromEnv()

	// Logger
	var loggerConfig zap.Config
	if *flagDevLog {
		loggerConfig = zap.NewDevelopmentConfig()
	} else {
		loggerConfig = zap.NewProductionConfig()
	}
	loggerConfig.Level = zap.NewAtomicLevelAt(*flagLogLevel)
	logger, err := loggerConfig.Build()
	if err != nil {
		panic(err.Error())
	}
	defer logger.Sync()
	log := zapr.NewLogger(logger)

	// Database connection
	dbConnection := db.NewDBConnection(nil,
		&dbConnectionParams,
		&log,
	)
	err = dbConnection.InitConnection()
	if err != nil {
		log.Error(err, "Failed to initialize DB connection")
		os.Exit(1)
	}
	defer dbConnection.Close()

	// API Server
	server := api.NewServer(*flagListenAddress,
		dbConnection,
		mailer,
		&log,
		flagCertPath,
		flagCertKeyPath,
	)
	ch := server.Start()
	err = <-ch
	if err != nil {
		panic(err.Error())
	}
}
