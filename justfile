default:
	@just --list

build:
	go install .

test: build
	maelstrom test -w broadcast --bin ~/go/bin/broadcast --node-count 1 --time-limit 20 --rate 10

test-multi-node: build
	maelstrom test -w broadcast --bin ~/go/bin/broadcast --node-count 5 --time-limit 30 --rate 20

test-fault-tolerance: build
	maelstrom test -w broadcast --bin ~/go/bin/broadcast --node-count 5 --time-limit 20 --rate 10 --nemesis partition

# Starts up the maelstrom server which shows detailed results.
results:
	maelstrom serve
