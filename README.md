# Xen Statsd

A simple go program that gets data from 
Xen and pushes it to a statsd server

# Install 

```bash
$ git clone https://github.com/syed/go-xen-statsd.git
$ cd go-xen-statsd
$ make 
$ cp xenstatsd-conf-sample.json /etc/xenstatsd.json
```

This generates a `xenstatsd` binary in the current directory. You
can run this from wherever you want

# Configure

The file `/etc/xenstatsd.json` stores the config in JSON format. 
Currently it supports a list of Xenserver hosts and the statsd
host where the data needs to be pushed. 
