package main

import (
	"context"
	"handler/function"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/openfaas/of-watchdog/config"

	"github.com/openfaas/of-watchdog/pkg"
)

const defaultTimeout = 10 * time.Second

func main() {

	os.Setenv("mode", "inproc")

	if _, ok := os.LookupEnv("exec_timeout"); !ok {
		os.Setenv("exec_timeout", defaultTimeout.String())
	}

	extendedTimeout := time.Duration(defaultTimeout + time.Millisecond*100)

	if _, ok := os.LookupEnv("read_timeout"); !ok {
		os.Setenv("read_timeout", extendedTimeout.String())
	}
	if _, ok := os.LookupEnv("write_timeout"); !ok {
		os.Setenv("write_timeout", extendedTimeout.String())
	}

	if v, ok := os.LookupEnv("healthcheck_interval"); !ok {
		os.Setenv("healthcheck_interval", os.Getenv("write_timeout"))
	} else {
		interval, _ := time.ParseDuration(v)
		if interval <= time.Millisecond*0 {
			os.Setenv("healthcheck_interval", "1ms")
		}
	}

	cfg, err := config.New(os.Environ())
	if err != nil {
		log.Fatalf("failed to parse watchdog config: %v", err)
	}

	log.Printf("Watchdog config: %+v\n", cfg)

	h := func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Header["X-Call-Id"]; !ok {
			r.Header.Set("X-Call-Id", uuid.New().String())
		}
		function.Handle(w, r)
	}

	cfg.SetHandler(h)
	cfg.OperationalMode = config.ModeInproc

	watchdog := pkg.NewWatchdog(cfg)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		s := <-sigs
		log.Printf("Signal received: %s", s.String())
		cancel()
	}()

	if err := watchdog.Start(ctx); err != nil {
		log.Fatalf("failed to start watchdog: %v", err)
	}
}
