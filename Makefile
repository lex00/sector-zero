.PHONY: build run release traces tidy clean

BINARY := sector-zero
GAME_DIR := ./game

build:
	go build -o $(BINARY) $(GAME_DIR)

release:
	go build -ldflags="-X main.Version=$(shell git describe --tags --always)" -o $(BINARY) $(GAME_DIR)

run: build
	./$(BINARY)

traces:
	@echo "Generating puzzle trace files..."
	cd ref && PYTHONPATH=../probes/python python3 bubble_sort.py > ../game/puzzles/data/01_bubble_sort.trace
	cd ref && PYTHONPATH=../probes/python python3 binary_search.py > ../game/puzzles/data/03_binary_search.trace
	cd ref && PYTHONPATH=../probes/python python3 merge_sort.py > ../game/puzzles/data/06_merge_sort.trace
	cd ref && PYTHONPATH=../probes/python python3 quick_sort.py > ../game/puzzles/data/07_quick_sort.trace
	@echo "Done."

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)
	rm -f /tmp/sz_run_*

test:
	go test ./...
