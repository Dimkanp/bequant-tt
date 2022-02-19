package app

import (
	"bequant-tt/cryptocompare/api/rest"
	"bequant-tt/cryptocompare/configuration"
	"bequant-tt/cryptocompare/repository"
	"bequant-tt/cryptocompare/service"
	"bequant-tt/pkg/runner"
)

type App struct {
	Rest       *rest.Rest
	Service    *service.Service
	Repository repository.Repository

	config configuration.Configuration
	runner runner.Runner
}

func New(config *configuration.Configuration) (*App, error) {
	var err error
	app := &App{config: *config}

	repo, err := repository.New(config.DB)
	if err != nil {
		return nil, err
	}

	err = repo.Migrate()
	if err != nil {
		return nil, err
	}

	app.Repository = repo

	cfg := &service.Config{
		Repository: app.Repository,
		Services:   &config.Services,
	}
	app.Service = service.New(cfg)

	app.Rest = rest.New(app.Service, &config.RestApi)

	app.runner = runner.New(app.Rest, runner.FromCloser(repo))

	return app, nil
}

func (app *App) Run() error {
	return app.runner.Run()
}

func (app *App) Stop() error {
	return app.runner.Stop()
}
