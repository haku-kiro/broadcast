default:
	@just --list

build:
	echo "todo: add the build path for you node"

test: build
	echo "Update this to point to the binary when you're done"
	maelstrom test -w broadcast --bin ~/go/bin/<node name> --node-count 1 --time-limit 20 --rate 10

# Starts up the maelstrom server which shows detailed results.
results:
	maelstrom serve
