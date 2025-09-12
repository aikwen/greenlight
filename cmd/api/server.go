package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// start a background goroutine
	go func() {
		quit := make(chan os.Signal, 1)
		// listen for incoming SIGINT and SIGTERM signals
		// and rely them to the quit channel
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		app.logger.PrintInfo("caught signal", map[string]string{
			"signal": s.String(),
		})

		os.Exit(0)
	}()

	app.logger.PrintInfo("starting server", map[string]string{
		"port": srv.Addr,
		"env":  app.config.env,
	})

	return srv.ListenAndServe()
}
