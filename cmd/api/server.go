package main

import (
	"context"
	"errors"
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

	shutdownError := make(chan error)

	// start a background goroutine
	go func() {
		quit := make(chan os.Signal, 1)
		// listen for incoming SIGINT and SIGTERM signals
		// and relay them to the quit channel
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		// create a context with a 5-second timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 1.if the graceful shutdown was successful Shutdown()
		// will return nil
		// 2.error which may happen because of a problem closing
		// the listeners
		// 3.context deadline is hit
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		// log a message to say that we're waiting for any background goroutines to
		// complete their tasks
		app.logger.PrintInfo("completing background tasks", map[string]string{
			"addr": srv.Addr,
		})

		app.wg.Wait()
		shutdownError <- nil

	}()

	app.logger.PrintInfo("starting server", map[string]string{
		"port": srv.Addr,
		"env":  app.config.env,
	})

	// calling Shutdown() on server will
	// immediately return an http.ErrServerClosed error.
	// if we see this error, it's actually a good thing
	// and an indication that the graceful shutdown has started
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})
	return nil
}
