.PHONY: linux darwin clean all

linux:
	env GOOS=linux GOARCH=amd64 go build -o builds/linux/mesos-fw

darwin:
	env GOOS=darwin GOARCH=amd64 go build -o builds/darwin/mesos-fw

all: linux darwin

clean:
	rm -rf builds/