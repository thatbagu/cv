BIN := cv-ssh-tui

.PHONY: build run clean

build:
	go build -o $(BIN) .

run: build
	./$(BIN)

clean:
	rm -f $(BIN)
