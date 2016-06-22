package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ctx struct {
	s *Settings

	restartC chan struct{}

	sync.Mutex
	sod   bool
	busy  bool
	hosts map[string]struct{}
}

func (c *ctx) logNoTime(format string, args ...interface{}) error {
	f, err := os.OpenFile(c.s.Logfile, os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if c.s.Verbose {
		fmt.Printf(format, args...)
	}

	_, err = fmt.Fprintf(f, format, args...)
	if err != nil {
		return err
	}

	return nil
}

func (c *ctx) log(format string, args ...interface{}) error {
	t := time.Now().Format("15:04:05.000 ")
	return c.logNoTime(t+format, args...)
}

func (c *ctx) writeHosts() error {
	c.log("Writing: %v\n", c.s.Target)

	f, err := os.Create(c.s.Target)
	if err != nil {
		return err
	}
	defer f.Close()

	for k := range c.hosts {
		_, err = fmt.Fprintf(f, "local-zone: \"%v\" redirect\n", k)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(f, "local-data: \"%v A 0.0.0.0\"\n", k)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *ctx) parseHosts(b []byte) error {
	buf := bytes.NewBuffer(b)

	i := 0
	for {
		line, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		i++

		if strings.HasPrefix(line, "#") {
			continue
		}

		a := strings.Split(line, " ")
		if len(a) < 2 {
			continue
		}

		c.hosts[strings.TrimSpace(a[1])] = struct{}{}
	}

	return nil
}

func (c *ctx) updateBackground(rerror *error) {
	var err error // err := is verboten in this function

	c.busy = true

	defer func() {
		if rerror != nil {
			*rerror = err
		}

		if err == nil && c.s.Update == false && c.sod == false {
			c.restartC <- struct{}{}
		}

		c.busy = false
	}()

	var f *os.File
	f, err = os.Open(c.s.Hosts)
	if err != nil {
		return
	}
	defer f.Close()

	br := bufio.NewReader(f)
	i := 0
	for {
		var uri string
		uri, err = br.ReadString('\n')
		if err == io.EOF {
			err = nil
			break
		}
		uri = strings.TrimSpace(uri)
		i++

		if strings.HasPrefix(uri, "#") {
			continue
		}

		c.log("Downloading: %v\n", uri)

		var body []byte
		body, err = downloadToMem(uri, false)
		if err != nil {
			return
		}

		err = c.parseHosts(body)
		if err != nil {
			return
		}
	}

	err = c.writeHosts()
	if err != nil {
		return
	}
}

func (c *ctx) update() error {
	c.Lock()
	defer c.Unlock()

	if c.busy {
		return fmt.Errorf("busy")
	}

	go c.updateBackground(nil)

	return nil
}

func (c *ctx) restart() error {
	args := strings.Split(c.s.Restart, " ")
	if len(args) < 1 {
		return fmt.Errorf("no restart command provided")
	}

	for k, v := range args {
		args[k] = strings.TrimSpace(v)
	}

	c.log("Restarting: %v\n", c.s.Restart)

	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	c.log("%v\n", string(out))

	return nil
}

func _main() error {
	var err error

	c := &ctx{
		hosts:    make(map[string]struct{}),
		restartC: make(chan struct{}),
		sod:      true,
	}
	c.s, err = parseSettings()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(c.s.Logfile), 0700)
	if err != nil {
		return err
	}

	// always run first
	c.updateBackground(&err)
	if err != nil {
		return err
	}

	if c.s.Update {
		return err // yep this is right
	}

	// restart if we aren't just updating file
	err = c.restart()
	if err != nil {
		return err
	}

	c.sod = false
	t := time.NewTicker(time.Second * time.Duration(c.s.Interval))
	for {
		select {
		case <-t.C:
			err = c.update()
			if err != nil {
				c.log("update failed: %v\n", err)
			}
		case <-c.restartC:
			err = c.restart()
			if err != nil {
				c.log("restart failed: %v\n", err)
			}
		}
	}

	return nil
}

func main() {
	err := _main()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
