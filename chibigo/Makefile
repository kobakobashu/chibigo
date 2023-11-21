CFLAGS=-std=c11 -g -static

PROJECT_NAME = chibigo

SRCS=$(wildcard *.go)

chibigo: $(SRCS)
	go build -o $(PROJECT_NAME)

test: chibigo
	./test.sh

clean:
	rm -f $(PROJECT_NAME) *.o *~ tmp*

.PHONY: test clean