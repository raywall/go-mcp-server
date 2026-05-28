.PHONY: start stop all

include assets/database.mk
include assets/tests.mk
include assets/mcp.mk

# Sobe os containers de banco de dados
start:
	@echo "Subindo Postgres e DynamoDB Local..."; \
	 docker-compose up -d; \
	 echo "Aguardando inicialização (5s)..."; \
	 sleep 5;

# Derruba a infraestrutura
stop:
	@docker-compose down -v;

# Fluxo completo
all: start seed bench