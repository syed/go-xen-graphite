# Xen Graphite

A simple go program that gets data from 
Xen and pushes it to a graphite server

# Install 

```bash
$ git clone https://github.com/syed/go-xen-graphite.git
$ cd go-xen-graphite
$ make 
$ cp xengraphite-conf-sample.json /etc/xengraphite.json
```

This generates a `xengraphite` binary in the current directory. You
can run this from wherever you want

# Configure

The file `/etc/xengraphite.json` stores the config in JSON format. 
Currently it supports a list of Xenserver hosts and the graphite
host where the data needs to be pushed. 
