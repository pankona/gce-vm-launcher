
all:
	find ./cmd -mindepth 1 -type d | xargs -I{} sh -c 'cd {}; go build'
