# ==========================================
# Testes de Inferência MCP
# ==========================================

.PHONY: build infer inspect

# Prepara e compila a aplicação localmente
build:
	@cd app; \
	 go build -o ../bin/mcp-server server.go

# 1. Teste de Inferência via CLI (simula o LLM via terminal)
# Requer o utilitário 'jq' instalado na máquina
infer: build
	@echo "Disparando payload de inferência (Tool Call) para o servidor MCP..."; \
	 echo '{"jsonrpc":"2.0","method":"tools/call","id":1,"params":{"name":"analyze_distributed_liquidation","arguments":{"cpf":"12345678900","amount":"450.50","date":"2026-05-10"}}}' \
	 | ./bin/mcp-server | grep '"result"' | jq -r '.result.content[0].text' | jq .

# 2. Inspetor Oficial do Protocolo MCP (UI Web)
# Requer Node.js/npx instalado na máquina
inspect: build
	@echo "Iniciando o MCP Inspector Oficial..."; \
	 echo "Acesse a URL gerada no terminal para testar as ferramentas interativamente."; \
	 npx -y @modelcontextprotocol/inspector ./bin/mcp-server;