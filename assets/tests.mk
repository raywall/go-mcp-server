.PHONY: tests

# Executa testes e benchmarks
bench:
	@echo "Executando benchmarks de performance e escalabilidade..."; \
	 cd app; \
	 go test -bench=. -benchmem -cpu=1,4,8 ./...

