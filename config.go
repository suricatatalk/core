package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
	"github.com/sebest/logrusly"
)

type AppConfig struct {
	Host string `default:"0.0.0.0"`
	Port string `default:"7070"`
	Name string `default:"core1"`
}

type MgoConfig struct {
	URI string `default:"127.0.0.1:27017"`
	DB  string `default:"surikata"`
}

type EtcdConfig struct {
	Endpoint string `default:"127.0.0.1:4001"`
}

// loadConfiguration loads the configuration of application
func loadConfiguration(app *AppConfig, mgo *MgoConfig, etcd *EtcdConfig) {
	err := envconfig.Process("core", app)
	if err != nil {
		log.Panicln(err)
	}
	err = envconfig.Process("mongodb", mgo)
	if err != nil {
		log.Panicln(err)
	}
	err = envconfig.Process("etcd", etcd)
	if err != nil {
		log.Panicln(err)
	}
	if len(os.Getenv(KeyLogly)) > 0 {
		hook := logrusly.NewLogglyHook(os.Getenv(KeyLogly),
			os.Getenv(KeyCoreHost),
			logrus.InfoLevel,
			os.Getenv(KeyCoreName))
		logrus.AddHook(hook)
	}
}
