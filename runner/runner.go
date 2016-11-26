package runner

import (
	"strings"
	"time"

	"chill/command"
)

type Runner interface {
	Path() string
	Patterns() []string
	Command() command.Command
	Start()
	Exit()
}

type runner struct {
	path     string
	patterns []string
	command  command.Command
	abort    chan struct{}
}

func NewRunner(path string, patterns []string, command command.Command) Runner {
	return &runner{path: path, patterns: patterns, command: command}
}

func (r *runner) Path() string {
	return r.path
}

func (r *runner) Patterns() []string {
	return r.patterns
}

func (r *runner) Command() command.Command {
	return r.command
}

func (r *runner) Start() {
	r.abort = make(chan struct{})

	changed, err := watch(r.path, r.abort)

	if err != nil {
		log.Fatalf("Failed to initialize watcher: %s", err.Error())
	}

	matched := match(changed, r.patterns)
	log.Info("Start watching......")

	r.command.Start(time.Millisecond * 200)
	for fp := range matched {
		files := gather(fp, matched, time.Millisecond*500)

		// Terminate previous running command
		r.command.Terminate(time.Second * 2)

		log.Infof("File changed: %s", strings.Join(files, ", "))

		// Run new command
		r.command.Start(time.Millisecond * 200)
	}
}

func (r *runner) Exit() {
	log.Info("Shutting down......")

	r.abort <- struct{}{}
	close(r.abort)
	r.command.Terminate(time.Second * 2)
}
