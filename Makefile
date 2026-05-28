.PHONY: start stop all

include assets/database.mk
include assets/tests.mk
include assets/mcp.mk

# Sobe os containers de banco de dados
start:
	@set -Eeuo pipefail; \
	 echo "Subindo Postgres e DynamoDB Local..."; \
	 docker-compose up -d; \
	 echo "Aguardando inicialização (5s)..."; \
	 sleep 5;

# Derruba a infraestrutura
stop:
	@set -Eeuo pipefail; \
	 docker-compose down -v;

# Fluxo completo
all: start import-rules seed bench