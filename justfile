binary := "sector-zero"
game_dir := "./game"

build:
    go build -o {{binary}} {{game_dir}}

run: build
    ./{{binary}}

release:
    go build -ldflags="-X main.Version=$(git describe --tags --always)" -o {{binary}} {{game_dir}}

traces:
    cd ref && python3 bubble_sort.py > ../game/puzzles/data/bubble_sort.trace
    cd ref && python3 binary_search.py > ../game/puzzles/data/binary_search.trace
    cd ref && python3 merge_sort.py > ../game/puzzles/data/merge_sort.trace
    cd ref && python3 quick_sort.py > ../game/puzzles/data/quick_sort.trace

tidy:
    go mod tidy

test:
    go test ./...

clean:
    rm -f {{binary}}
    rm -f /tmp/sz_run_*
