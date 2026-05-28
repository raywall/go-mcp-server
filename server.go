package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/lib/pq"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Configurações do Banco de Dados de Regras
const dbConnStr = "postgres://postgres:postgres@localhost:5432/rules_db?sslmode=disable"

// Rule agrupa os dados de uma regra de negócio, incluindo sua ordem lógica de avaliação
type Rule struct {
	ID             int    `json:"rule_id"`
	Domain         string `json:"domain"`
	Description    string `json:"description"`
	ExecutionOrder int    `json:"execution_order"`
}

// TransactionContext agora recebe as regras de forma estruturada e ordenada
type TransactionContext struct {
	RequestedCPF   string `json:"cpf"`
	ExpectedAmount string `json:"expected_amount"`
	RequestedDate  string `json:"requested_date"`
	Rules          []Rule `json:"business_rules"` // Alterado de []string para []Rule
	APIData        struct {
		Custody    string `json:"custody_status"`
		Consignado string `json:"consignado_status"`
		Payments   string `json:"payment_received"`
	} `json:"api_data"`
	Logs struct {
		Baixa         string `json:"baixa_log"`
		Ressarcimento string `json:"ressarcimento_log"`
	} `json:"system_logs"`
}

func main() {
	// 1. Instancia o servidor MCP
	s := server.NewMCPServer(
		"Consignado-Liquidacao-MCP",
		"1.0.0",
	)

	// 2. Registra a ferramenta (Tool) que o LLM poderá chamar
	tool := mcp.NewTool("analyze_liquidation",
		mcp.WithDescription("Analisa o fluxo de baixa e ressarcimento de uma parcela de crédito consignado, buscando dados em APIs, bancos e logs."),
		mcp.WithString("cpf", mcp.Required(), mcp.Description("CPF do cliente (apenas números)")),
		mcp.WithString("amount", mcp.Required(), mcp.Description("Valor da parcela esperada")),
		mcp.WithString("date", mcp.Required(), mcp.Description("Data de vencimento da parcela (YYYY-MM-DD)")),
	)

	// 3. Define o handler da ferramenta
	s.AddTool(tool, analyzeLiquidationHandler)

	// 4. Inicia o servidor em modo STDIO (padrão de comunicação do MCP com LLMs)
	log.Println("Iniciando MCP Server de Consignado via STDIO...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Erro no servidor: %v", err)
	}
}

// analyzeLiquidationHandler é a função invocada pelo LLM quando ele decide usar a ferramenta
func analyzeLiquidationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 1. Faz o type assertion de 'any' para 'map[string]interface{}'
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Argumentos inválidos ou ausentes no request MCP"), nil
	}

	// 2. Extrai os valores com segurança
	cpf, _ := args["cpf"].(string)
	amount, _ := args["amount"].(string)
	date, _ := args["date"].(string)

	// 3. Popula o contexto base (Agora usando a variável date!)
	txContext := TransactionContext{
		RequestedCPF:   cpf,
		ExpectedAmount: amount,
		RequestedDate:  date,
	}

	// 4. Captura regras do Banco de Dados (Postgres)
	rules, err := fetchBusinessRules()
	if err != nil {
		// Injeta uma regra "falsa" do tipo struct Rule para relatar o erro de conexão ao LLM
		txContext.Rules = []Rule{
			{
				ID:             999,
				Domain:         "system_error",
				Description:    "Erro ao carregar regras do banco de dados: " + err.Error(),
				ExecutionOrder: 0,
			},
		}
	} else {
		txContext.Rules = rules
	}

	// 5. Simula requisições a APIs REST usando as chaves (incluindo a date)
	txContext.APIData.Custody = mockRestCall("API_CUSTODIA", cpf, date) // <- date sendo usada aqui
	txContext.APIData.Consignado = mockRestCall("API_CONSIGNADO", cpf, amount)
	txContext.APIData.Payments = mockRestCall("API_PAGAMENTOS", cpf, amount)

	// 6. Simula leitura de Logs
	txContext.Logs.Baixa = mockLogSearch("worker-baixa", cpf, amount)
	txContext.Logs.Ressarcimento = mockLogSearch("worker-ressarcimento", cpf, amount)

	// 7. Serializa o resultado enriquecido para devolver ao LLM
	resultJSON, _ := json.MarshalIndent(txContext, "", "  ")

	return mcp.NewToolResultText(string(resultJSON)), nil
}

// fetchBusinessRules extrai as regras do banco ordenadas pela prioridade lógica
func fetchBusinessRules() ([]Rule, error) {
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// O ORDER BY execution_order ASC é o coração desta alteração
	query := `
		SELECT id, domain, description, execution_order 
		FROM rules 
		WHERE domain IN ('baixa', 'ressarcimento', 'consignado')
		ORDER BY execution_order ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []Rule
	for rows.Next() {
		var r Rule
		if err := rows.Scan(&r.ID, &r.Domain, &r.Description, &r.ExecutionOrder); err == nil {
			rules = append(rules, r)
		}
	}
	return rules, nil
}

// mockRestCall simula uma chamada HTTP para as APIs de microserviços
func mockRestCall(service, key1, key2 string) string {
	// Simula latência de rede
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))

	switch service {
	case "API_CUSTODIA":
		return fmt.Sprintf("Contrato ativo localizado para cliente %s na data %s.", key1, key2)
	case "API_PAGAMENTOS":
		// Simula um cenário onde o RH repassou a maior
		return fmt.Sprintf("Repasse de averbação recebido via folha: R$ %s (Status: PROCESSADO)", key2)
	default:
		return "Dados consistentes."
	}
}

// mockLogSearch simula uma busca de logs distribuídos
func mockLogSearch(index, cpf, amount string) string {
	if index == "worker-baixa" {
		return fmt.Sprintf("[WARN] Heurística aplicada: CPF %s. Baixa parcial executada. Valor recebido difere do contrato.", cpf)
	}
	return fmt.Sprintf("[INFO] Nenhum evento de ressarcimento encontrado na janela temporal de 48h para R$ %s", amount)
}
