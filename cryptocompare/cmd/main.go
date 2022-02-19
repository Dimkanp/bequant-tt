package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bequant-tt/cryptocompare"
	"bequant-tt/cryptocompare/configuration"

	"github.com/spf13/viper"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "/app/data/config.json", "")
	flag.Parse()

	cfg, err := configuration.ProcessConfig(getConfig(configPath))
	if err != nil {
		err = fmt.Errorf("process configuration error: %s", err)
		log.Fatal(err)
	}

	a, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	errChan := make(chan error)
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		errChan <- a.Run()
	}()

	select {
	case err := <-errChan:
		log.Println("Application error: ", err)
	case _ = <-termChan:
		err := a.Stop()
		if err != nil {
			log.Println("Error on stopping application: ", err)
		}
	}
}

func getConfig(path string) configuration.InAppConfiguration {
	if path != "" {
		viper.SetConfigFile(path)
	} else {
		viper.SetConfigName("/app/data/config.json")
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Println(err)
	}

	var cfg configuration.InAppConfiguration

	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
