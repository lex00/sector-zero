.PHONY: build run traces tidy clean

BINARY := sector-zero
GAME_DIR := ./game

build:
	go build -o $(BINARY) $(GAME_DIR)

run: build
	./$(BINARY)

traces:
	@echo "Generating puzzle trace files..."
	cd ref && python3 bubble_sort.py > ../game/puzzles/data/bubble_sort.trace
	cd ref && python3 binary_search.py > ../game/puzzles/data/binary_search.trace
	cd ref && python3 merge_sort.py > ../game/puzzles/data/merge_sort.trace
	cd ref && python3 quick_sort.py > ../game/puzzles/data/quick_sort.trace
	@echo "Done."

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)
	rm -f /tmp/sz_run_*

test:
	go test ./...
