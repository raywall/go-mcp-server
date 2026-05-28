.PHONY: seed import-rules

# Cria tabelas no DynamoDB e injeta a massa de teste (Regras e Dados)
seed:
	@echo "--- Preparando Tabela de Baixas (Liquidations) ---"; \
	 aws dynamodb create-table --endpoint-url http://localhost:8000 \
		--table-name liquidations \
		--attribute-definitions AttributeName=cpf,AttributeType=S AttributeName=sk_date,AttributeType=S \
		--key-schema AttributeName=cpf,KeyType=HASH AttributeName=sk_date,KeyType=RANGE \
		--billing-mode PAY_PER_REQUEST --region us-east-1 > /dev/null 2>&1 || true; \
	
	 echo "--- Injetando Registro de Falha Simulado ---"; \
	 aws dynamodb put-item --endpoint-url http://localhost:8000 --region us-east-1 \
		--table-name liquidations \
		--item '{"cpf": {"S": "12345678900"}, "sk_date": {"S": "2026-05-10"}, "status": {"S": "RETIDO_PREV_INCONSISTENCIA"}}'; \
	 echo "✅ Dado de teste inserido com sucesso para o CPF 12345678900.";

# Popula o DynamoDB lendo os YAMLs do repositório
import-rules:
	@echo "Lendo YAMLs e gravando regras no DynamoDB..."; \
	 cd app; \
	 go run server.go --mode=import;