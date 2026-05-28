.PHONY: seed import-rules

# Cria tabelas no DynamoDB e injeta a massa de teste (Regras e Dados)
seed:
	@set -Eeuo pipefail; \
	 echo "--- Preparando Tabela de Baixas (Liquidations) ---"; \
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
	@set -Eeuo pipefail; \
	 echo "--- Preparando Tabela de Regras (business_rules) ---"; \
	 aws dynamodb create-table --endpoint-url http://localhost:8000 \
		--table-name business_rules \
		--attribute-definitions AttributeName=domain,AttributeType=S AttributeName=execution_order,AttributeType=N \
		--key-schema AttributeName=domain,KeyType=HASH AttributeName=execution_order,KeyType=RANGE \
		--billing-mode PAY_PER_REQUEST --region us-east-1 > /dev/null 2>&1 || true; \
	 echo "--- Lendo YAMLs e gravando regras no DynamoDB ---"; \
	 cd app; \
	 go run . --mode import;