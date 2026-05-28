CREATE TABLE rules (
    id SERIAL PRIMARY KEY,
    domain VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    execution_order INT NOT NULL
);

-- Inserindo as regras com a ordem lógica de avaliação
INSERT INTO rules (domain, description, execution_order) VALUES
('consignado', 'REGRA_01: Contratos em situação de sinistro ou óbito bloqueiam qualquer rotina, abortar processo imediatamente.', 10),
('baixa', 'REGRA_02: A baixa da parcela só pode ocorrer se o valor recebido da folha for maior ou igual ao valor em custódia.', 20),
('baixa', 'REGRA_03: Se não houver ID de repasse, correlacionar via CPF, CNPJ e valor exato (margem de 2 centavos).', 30),
('ressarcimento', 'REGRA_04: Se o repasse for maior que a parcela, a diferença dispara um evento de estorno na conta em D+1.', 40);