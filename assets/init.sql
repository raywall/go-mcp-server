CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    cpf VARCHAR(14) NOT NULL,
    amount NUMERIC(10, 2) NOT NULL,
    payment_date DATE NOT NULL,
    status VARCHAR(50) NOT NULL
);

-- Inserindo massa de dados transacional (Pagamentos recebidos)
INSERT INTO payments (cpf, amount, payment_date, status) VALUES
('12345678900', 450.50, '2026-05-10', 'PROCESSADO_VIA_FOLHA'),
('98765432100', 1200.00, '2026-05-12', 'RETIDO_DIVERGENCIA');