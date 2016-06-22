# adfuccant

Hate ads?  So do I.

This tool downloads hosts files from provided URLs and transforms them into
unbound configuration format.  Once complete it restarts the unbound server.

The idea of this tool is to block ads, porn, phishing etc at wherever you run
your edge DNS server.

# Usage
```
  -hosts string
        file that contains URL' to hosts files (default "hosts.txt")
  -interval int
        interval in seconds between hosts files updates (default 3600)
  -logfile string
        log file (default "~/.adfuccant/adfuccant.log")
  -restart string
        command to restart unbound (default "/etc/rc.d/unbound reload")
  -target string
        target file (default "/var/unbound/etc/local-blocking-data.conf")
  -update
        update target file only and exit
  -verbose
        print logging information to screen
```

# Unbound

Add the following to your unbound.conf file:
```
server:
        access-control: 0.0.0.0/8 allow
        include: /var/unbound/etc/local-blocking-data.conf
```

# Example in Bitrig

To start adfuccant on boot add something along these lines to /etc/rc.local:
```
adfuccant --hosts /root/hosts.txt &
```
