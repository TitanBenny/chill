package main

import (
	"os"
	"os/signal"
	"path/filepath"

	"chill/command"
	"chill/config"
	"chill/runner"
)

func main() {
	conf := config.GetConfigs()

	abspath, _ := filepath.Abs(conf.Directory)

	patterns := conf.Patterns
	cmd := command.New(conf.Command)

	r := runner.NewRunner(abspath, patterns, cmd)

	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch, os.Interrupt)

		<-ch
		r.Exit()
	}()
	r.Start()
}
