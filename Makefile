all: xengraphite

xengraphite:
	go build app/xengraphite.go

clean:
	rm xengraphite

