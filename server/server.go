// Package routes is a commen place to put all applicatioin routes.
// In order to easy setup routes for application and testing.
package server

import (
	"os"
	"os/signal"
	"time"
	"context"
	"fmt"
	golog "log"
	"net/http"

	"github.com/go-chi/valve"
	"github.com/theplant/appkit/log"
)

type Config struct {
	Addr string `default:":9800"`
}

func newServer(config Config, logger log.Logger, valv *valve.Valve) *http.Server {
	//valv := valve.New()

	server := http.Server{
		Addr:     config.Addr,
		ErrorLog: golog.New(log.LogWriter(logger.Error()), "", golog.Llongfile),
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	//signal.Notify(c, syscall.SIGTERM)
	//signal.Notify(c, syscall.SIGINT)
	go func() {
		for range c {
			// sig is a ^C, handle it
			logger.Info().Log("shutting down..")

			// first valv
			valv.Shutdown(20 * time.Second)

			// create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			// start http shutdown
			server.Shutdown(ctx)

			// verify, in worst case call cancel via defer
			select {
			case <-time.After(21 * time.Second):
				logger.Error().Log("not all connections done")
			case <-ctx.Done():

			}
		}
	}()
	return &server
}

func ListenAndServe(config Config, logger log.Logger, handler http.Handler, valv *valve.Valve) {

	logger = logger.With("during", "server.ListenAndServe")
	s := newServer(config, logger, valv)
	s.Handler = handler

	logger.Info().Log(
		"addr", config.Addr,
		"msg", fmt.Sprintf("HTTP server listening on %s", config.Addr),
	)
	//s.ListenAndServe()
	if err := s.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed{
			logger.Error().Log(
				"msg", fmt.Sprintf("Error in ListenAndServe: %v", err.Error()),
				"err", err,
			)
		}
	}
}
