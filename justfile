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

tail-log:
	@tail --follow /tmp/log.json | jq '.'

search-log TERM:
	@cat /tmp/log.json | rg -i "{{TERM}}" | jq '.'

test: build clear-logs
	maelstrom test -w broadcast --bin ~/go/bin/broadcast --node-count 1 --time-limit 20 --rate 10

test-multi-node: build clear-logs
	maelstrom test -w broadcast --bin ~/go/bin/broadcast --node-count 5 --time-limit 20 --rate 10

test-fault-tolerance: build clear-logs
	maelstrom test -w broadcast --bin ~/go/bin/broadcast --node-count 5 --time-limit 20 --rate 10 --nemesis partition

test-efficiency-a: build clear-logs
	maelstrom test -w broadcast --bin ~/go/bin/broadcast --node-count 25 --time-limit 20 --rate 100 --latency 100

test-efficiency-b: build clear-logs
	maelstrom test -w broadcast --bin ~/go/bin/broadcast --node-count 25 --time-limit 20 --rate 100 --latency 100 --nemesis partition

# Starts up the maelstrom server which shows detailed results.
results:
	maelstrom serve
