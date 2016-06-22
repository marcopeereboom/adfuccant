package main

import (
	"flag"

	"github.com/mitchellh/go-homedir"
)

type Settings struct {
	Interval int    // time between updates
	Hosts    string // hosts file to download
	Logfile  string // logfile
	Restart  string // command to restart unbound after updating
	Target   string // resulting unbound file
	Update   bool   // update file only
	Verbose  bool   // loudnes
}

const (
	defaultHosts   = "hosts.txt"
	defaultTarget  = "/var/unbound/local-blocking-data.conf"
	defaultLog     = "~/.adfuccant/adfuccant.log"
	defaultRestart = "/etc/rc.d/unbound reload"
)

func parseSettings() (*Settings, error) {
	s := Settings{}

	interval := flag.Int("interval", 3600,
		"interval in seconds between hosts files updates (default 3600)")
	hosts := flag.String("hosts", defaultHosts,
		"file that contains URL' to hosts files")
	logfile := flag.String("logfile", defaultLog, "log file")
	restart := flag.String("restart", defaultRestart,
		"command to restart unbound")
	target := flag.String("target", defaultTarget, "target file")
	update := flag.Bool("update", false, "update target file only and exit")
	verbose := flag.Bool("verbose", false,
		"print logging information to screen")
	flag.Parse()

	s.Interval = *interval
	s.Hosts = *hosts
	lf, err := homedir.Expand(*logfile)
	if err != nil {
		return nil, err
	}
	s.Logfile = lf
	s.Restart = *restart
	s.Target = *target
	s.Update = *update
	s.Verbose = *verbose

	return &s, nil
}
