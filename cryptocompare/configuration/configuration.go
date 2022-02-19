package configuration

import (
	"time"

	global "bequant-tt/configuration"
	"bequant-tt/cryptocompare/api/rest"
	"bequant-tt/cryptocompare/service"
)

// InAppConfiguration using to read in application configuration from file
// Looks like json config file to make check fields easier
type InAppConfiguration struct {
	RestApi struct {
		Address string
	}
	DB struct {
		UserName string
		Password string
		Host     string
		Port     int
		DBName   string
		SSL      string
	}
	Services struct {
		Scheduler struct {
			RefreshTime int // in seconds
		}
	}
}

type Configuration struct {
	RestApi  rest.Config
	DB       global.DBConfig
	Services service.ServicesConfig
}

func ProcessConfig(c InAppConfiguration) (_ *Configuration, err error) {
	options := &Configuration{
		RestApi: rest.Config{
			Address: c.RestApi.Address,
		},
		DB: global.DBConfig{
			UserName: c.DB.UserName,
			Password: c.DB.Password,
			Host:     c.DB.Host,
			Port:     c.DB.Port,
			DBName:   c.DB.DBName,
			SSL:      c.DB.SSL,
		},
		Services: service.ServicesConfig{
			Scheduler: &service.SchedulerConfig{
				RefreshTime: time.Second * time.Duration(c.Services.Scheduler.RefreshTime),
			},
		},
	}

	return options, nil
}
