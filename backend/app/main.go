package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/logutils"
	"github.com/jessevdk/go-flags"
	"github.com/theshamuel/data-sec-exchanger/backend/app/rest"
	"github.com/theshamuel/test_3p"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var opts struct {
	Debug bool `long:"debug" env:"DEBUG" description:"enable debug mode"`
}

var version = "unknown"

func main() {
	log.Printf("[INFO] Starting XXXX version:%s ...\n", version)
	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	p.SubcommandsOptional = true
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			log.Printf("[ERROR] cli error: %v", err)
		}
		os.Exit(2)
	}
	setupLogLevel(opts.Debug)
	log.Printf("[DEBUG] app options: %+v", opts)

	err := run()
	if err != nil {
		log.Fatalf("[ERROR] XXX application failed, %v", err)
	}
	test_3p.Add(1, 2)
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if r := recover(); r != nil {
			log.Printf("[WARN] run time panic:\n%v", r)
			panic(r)
		}
		// catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] interrupt signal")
		cancel()
	}()

	err := makeRestAPI(ctx, 3000)
	if err != nil {
		return fmt.Errorf("failed to make providers: %w", err)
	}
	return err
}

func makeRestAPI(ctx context.Context, port int) error {
	rest := rest.Rest{
		Version: version,
		URI:     "http://localhost",
	}
	rest.Run(ctx, port)
	return nil
}

func setupLogLevel(debug bool) {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("INFO"),
		Writer:   os.Stdout,
	}
	log.SetFlags(log.Ldate | log.Ltime)

	if debug {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
		filter.MinLevel = logutils.LogLevel("DEBUG")
	}
	log.SetOutput(filter)
}

func getStackTrace() string {
	maxSize := 7 * 1024 * 1024
	stacktrace := make([]byte, maxSize)
	length := runtime.Stack(stacktrace, true)
	if length > maxSize {
		length = maxSize
	}
	return string(stacktrace[:length])
}

func init() {
	sigChan := make(chan os.Signal)
	go func() {
		for range sigChan {
			log.Printf("[INFO] Singal QUITE is cought , stacktrace [\n%s", getStackTrace())
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)
}
