CFLAGS=-std=c11 -g -static

PROJECT_NAME = chibigo

chibigo:
	go build -o $(PROJECT_NAME)

test: chibigo
	./test.sh

clean:
	rm -f $(PROJECT_NAME) *.o *~ tmp*

.PHONY: test clean