.PHONY: build start stop test-mcp test bench profile all

# Prepara e compila a aplicação localmente
build:
	go build -o bin/mcp-server server.go

# Sobe o banco de dados (regras)
start:
	docker-compose up -d db
	@echo "Aguardando o banco subir..."
	sleep 3

# Derruba a infra
stop:
	docker-compose down -v

# Dispara uma simulação crua passando um JSON-RPC pela entrada padrão (STDIO)
# Isso simula o comportamento que o Claude/Gemini teria ao chamar a ferramenta
test-mcp: build
	@echo "Simulando requisição do LLM para a ferramenta analyze_liquidation:"
	@echo '{"jsonrpc":"2.0","method":"tools/call","id":1,"params":{"name":"analyze_liquidation","arguments":{"cpf":"12345678900","amount":"450.50","date":"2026-05-10"}}}' | ./bin/mcp-server

# Executa os testes unitários padrão
test:
	go test -v ./...

# Executa os benchmarks detalhando consumo de memória (-benchmem)
bench:
	@echo "Executando Benchmarks de Performance e Escalabilidade..."
	go test -bench=. -benchmem -cpu=1,4,8 ./...

# Gera arquivos de profile para análise de gargalos (CPU e Memória) via pprof
profile:
	go test -bench=BenchmarkAnalyzeLiquidationParallel -cpuprofile=cpu.prof -memprofile=mem.prof ./...
	@echo "Profiles gerados. Use 'go tool pprof cpu.prof' para analisar."

# 1. Executa benchmarks e salva em texto puro (Standard Output)
bench-txt:
	@echo "Executando benchmarks (Text Mode)..."
	go test -bench=. -benchmem -cpu=1,4,8 ./... > bench_results.txt
	@echo "✅ Resultados salvos em bench_results.txt"

# 2. Executa benchmarks e salva em JSON (Ideal para CI/CD e Observabilidade)
bench-json:
	@echo "Executando benchmarks (JSON Mode)..."
	go test -bench=. -benchmem -cpu=1,4,8 -json ./... > bench_results.json
	@echo "✅ Resultados estruturados salvos em bench_results.json"

# 3. Instala e usa o benchstat para formatar os resultados de forma tabular 
# Ele depende do arquivo txt gerado no passo 1
bench-stat: bench-txt
	@echo "Formatando resultados para leitura humana..."
	@go install golang.org/x/perf/cmd/benchstat@latest
	@$(shell go env GOPATH)/bin/benchstat bench_results.txt

# Executa todos os testes, benchmarks e gera profiles
all: start test-mcp bench bench-txt bench-json bench-stat stop