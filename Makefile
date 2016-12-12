all: xenstatsd

xenstatsd:
	go build app/xenstatsd.go

clean:
	rm xenstatsd

