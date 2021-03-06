package main

import (
	"context"
	"github.com/bearname/url-short/internal/short/app"
	"github.com/bearname/url-short/internal/short/infrastructure"
	"github.com/bearname/url-short/internal/short/infrastructure/postgres"
	"github.com/bearname/url-short/internal/short/infrastructure/router"
	"github.com/bearname/url-short/internal/short/infrastructure/transport"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

func main() {
	//runtime.GOMAXPROCS(4)
	//log.SetFormatter(&log.JSONFormatter{})
	//file, err := os.OpenFile("short.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	//if err == nil {
	//	log.SetOutput(file)
	//	defer func(file *os.File) {
	//		err := file.Close()
	//		if err != nil {
	//			log.Error(err)
	//		}
	//	}(file)
	//}

	//conf, err := ParseConfig()
	//if err != nil {
	//	log.Fatal("Default settings" + err.Error())
	//}

	conf := Config{
		ServeRestAddress: ":8000",
		DbAddress:        "localhost:5432",
		DbName:           "urlshort",
		DbUser:           "postgres",
		DbPassword:       "postgres",
		MaxConnections:   10,
		AcquireTimeout:   1,
	}

	connector, err := getConnector(conf)

	if err != nil {
		log.Fatal(err.Error())
	}
	pool, err := newConnectionPool(connector)

	if err != nil {
		log.Fatal(err.Error())
	}

	server := infrastructure.Server{}
	killSignalChan := server.GetKillSignalChan()
	repository := postgres.NewUrlRepository(pool)
	service := app.NewUrlService(repository)
	controller := transport.NewUrlController(service)
	handler := router.Router(controller)

	log.WithFields(log.Fields{"url": conf.ServeRestAddress}).Info("starting the server")

	srv := server.StartServer(conf.ServeRestAddress, handler)

	server.WaitForKillSignal(killSignalChan)
	err = srv.Shutdown(context.Background())
	if err != nil {
		log.Error(err)
		return
	}
}

func getConnector(config Config) (pgx.ConnPoolConfig, error) {
	databaseUri := "postgres://" + config.DbUser + ":" + config.DbPassword + "@" + config.DbAddress + "/" + config.DbName
	log.Info("databaseUri: " + databaseUri)
	pgxConnConfig, err := pgx.ParseURI(databaseUri)
	if err != nil {
		return pgx.ConnPoolConfig{}, errors.Wrap(err, "failed to parse database URI from environment variable")
	}
	pgxConnConfig.Dial = (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 5 * time.Minute}).Dial
	pgxConnConfig.RuntimeParams = map[string]string{
		"standard_conforming_strings": "on",
	}
	pgxConnConfig.PreferSimpleProtocol = true

	return pgx.ConnPoolConfig{
		ConnConfig:     pgxConnConfig,
		MaxConnections: config.MaxConnections,
		AcquireTimeout: time.Duration(config.AcquireTimeout) * time.Second,
	}, nil
}

func newConnectionPool(config pgx.ConnPoolConfig) (*pgx.ConnPool, error) {
	return pgx.NewConnPool(config)
}
