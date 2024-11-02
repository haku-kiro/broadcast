default:
	@just --list

build:
	go install .

test: build
	maelstrom test -w broadcast --bin ~/go/bin/broadcast --node-count 1 --time-limit 20 --rate 10

heavy-test: build
	maelstrom test -w broadcast --bin ~/go/bin/broadcast --node-count 2 --time-limit 20 --rate 10

# Starts up the maelstrom server which shows detailed results.
results:
	maelstrom serve
