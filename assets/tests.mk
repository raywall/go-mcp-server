.PHONY: tests

# Executa testes e benchmarks
bench:
	@go test -bench=. -benchmem -cpu=1,4,8 ./...

