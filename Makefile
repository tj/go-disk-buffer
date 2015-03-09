
test:
	@go test -cover

bench:
	@go test -bench=. -cpu 1,2,4,8 -run Benchmark

.PHONY: bench test