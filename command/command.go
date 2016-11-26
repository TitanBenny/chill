package command

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	colorlog "chill/log"
)

var log = colorlog.NewLog()

// Command is interface contains the String, Strar, Terminate method
// An interface{} type is a type that could be any value
type Command interface {
	String() string
	Start(delay time.Duration)
	Terminate(wait time.Duration)
}

// comand is a type that will be implement Command interface
type command struct {
	name string
	args []string
	cmd  *exec.Cmd

	// mutex lock
	mutex *sync.Mutex
	exit  chan struct{}
}

// factory mode
func New(cmdstrings []string) Command {
	if len(cmdstrings) == 0 {
		return Empty()
	}

	name := cmdstrings[0]
	// cmdstrings[1:] means [1,n)
	args := cmdstrings[1:]

	return &command{
		name:  name,
		args:  args,
		cmd:   nil,
		mutex: &sync.Mutex{},
		exit:  nil,
	}
}

// String
func (c *command) String() string {
	return fmt.Sprintf("%s %s", c.name, strings.Join(c.args, " "))
}

func (c *command) Start(delay time.Duration) {
	time.Sleep(delay) // delay for a while to avoid start frequently

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// ProcessState contains information about an exited process,
	// avaliable after a call to Wait or Run
	if c.cmd != nil && c.cmd.ProcessState.Exited() {
		log.Fatal("Failed to start command: previous command hasn't exit.")
	}
	cmd := exec.Command(c.name, c.args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err := cmd.Start()

	exit := make(chan struct{})

	if err != nil {
		log.Fatalf("Failed: %s", err.Error())
	} else {
		c.cmd = cmd
		c.exit = exit

		go func() {
			defer func() {
				exit <- struct{}{}
				close(exit)
			}()

			cmd.Wait()

			if cmd.ProcessState.Success() {
				log.Info("application has start successful!")
			} else {
				log.Info("terminated.")
			}
		}()
	}
}

func (c *command) Terminate(wait time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	defer func() {
		c.cmd = nil
	}()

	if c.cmd == nil {
		return
	}

	if c.cmd.ProcessState != nil && c.cmd.ProcessState.Exited() {
		return
	}

	log.Info("try to make application stopping......")
	// Try to stop the process by sending a SIGINT sinnal.
	if err := c.kill(syscall.SIGINT); err != nil {
		log.Fatalf("Failed to terminate process with interrupt: %s", err.Error())
	}

	for {
		select {
		case <-c.exit:
			return
		case <-time.After(wait):
			log.Info("-Killing process")
			c.kill(syscall.SIGTERM)
		}
	}
}

func (c *command) kill(sig syscall.Signal) error {
	cmd := c.cmd

	if pgid, err := syscall.Getpgid(cmd.Process.Pid); err == nil {
		return syscall.Kill(-pgid, sig)
	} else {
		return err
	}
}
