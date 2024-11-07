default:
	@just --list

build:
	go install .

clear-logs:
	rm -rf /tmp/log.json

error-logs:
	@cat /tmp/log.json | rg 'err' | jq '.'

logs:
	@cat /tmp/log.json | jq '.'

test: build clear-logs
	maelstrom test -w broadcast --bin ~/go/bin/broadcast --node-count 1 --time-limit 20 --rate 10

quick-test: build clear-logs
	# Testing by creating a network partition
	maelstrom test -w broadcast --bin ~/go/bin/broadcast --node-count 2 --time-limit 10 --rate 10 --nemesis partition

test-multi-node: build clear-logs
	maelstrom test -w broadcast --bin ~/go/bin/broadcast --node-count 5 --time-limit 30 --rate 20

test-fault-tolerance: build clear-logs
	maelstrom test -w broadcast --bin ~/go/bin/broadcast --node-count 5 --time-limit 20 --rate 10 --nemesis partition

# Starts up the maelstrom server which shows detailed results.
results:
	maelstrom serve
