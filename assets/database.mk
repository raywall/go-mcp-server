.PHONY: seed

# Cria tabelas no DynamoDB e injeta a massa de teste (Regras e Dados)
seed:
	@echo "Criando tabela de Regras no DynamoDB..."; \
	 aws dynamodb create-table --endpoint-url http://localhost:8000 --table-name business_rules \
		--attribute-definitions AttributeName=domain,AttributeType=S AttributeName=execution_order,AttributeType=N \
		--key-schema AttributeName=domain,KeyType=HASH AttributeName=execution_order,KeyType=RANGE \
		--billing-mode PAY_PER_REQUEST --region us-east-1 > /dev/null 2>&1 || true; \

	 echo "Criando tabela de Baixas (Dados) no DynamoDB..."; \
	 aws dynamodb create-table --endpoint-url http://localhost:8000 --table-name liquidations \
		--attribute-definitions AttributeName=cpf,AttributeType=S AttributeName=sk_date,AttributeType=S \
		--key-schema AttributeName=cpf,KeyType=HASH AttributeName=sk_date,KeyType=RANGE \
		--billing-mode PAY_PER_REQUEST --region us-east-1 > /dev/null 2>&1 || true; \

	 echo "Injetando Regras de Negócio (DynamoDB)..."; \
	 aws dynamodb put-item --endpoint-url http://localhost:8000 --table-name business_rules \
		--item '{"domain": {"S": "baixa"}, "execution_order": {"N": "10"}, "description": {"S": "Validar conciliação de centavos no relacional antes da baixa"}}' --region us-east-1; \
	 aws dynamodb put-item --endpoint-url http://localhost:8000 --table-name business_rules \
		--item '{"domain": {"S": "baixa"}, "execution_order": {"N": "20"}, "description": {"S": "Se divergente, acionar API de inconsistências"}}' --region us-east-1; \

	 echo "Injetando Dados de Baixa (DynamoDB)..."; \
	 aws dynamodb put-item --endpoint-url http://localhost:8000 --table-name liquidations \
		--item '{"cpf": {"S": "12345678900"}, "sk_date": {"S": "2026-05-10"}, "status": {"S": "PENDENTE_INTEGRACAO"}}' --region us-east-1; \

	 echo "✅ Massa de dados (Postgres e Dynamo) pronta para consumo.";