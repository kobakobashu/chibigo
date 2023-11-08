CFLAGS=-std=c11 -g -static

chibigo: main.go
	go build -o chibigo main.go

test: chibigo
	./test.sh

clean:
	rm -f chibigo *.o *~ tmp*

.PHONY: test clean