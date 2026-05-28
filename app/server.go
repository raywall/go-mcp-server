package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	_ "github.com/lib/pq"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// App encapsula as conexões de I/O
type App struct {
	DynamoClient *dynamodb.Client
	PostgresDB   *sql.DB
}

// ==========================================
// Abstrações de Domínio (Agrupamento de Dados)
// ==========================================

// Rule abstrai a regra de negócio vinda do DynamoDB
type Rule struct {
	Domain         string `dynamodbav:"domain" json:"domain"`
	ExecutionOrder int    `dynamodbav:"execution_order" json:"execution_order"`
	Description    string `dynamodbav:"description" json:"description"`
}

// LiquidationRecord abstrai os dados operacionais da baixa (DynamoDB)
type LiquidationRecord struct {
	CPF    string `dynamodbav:"cpf" json:"cpf"`
	SKDate string `dynamodbav:"sk_date" json:"sk_date"`
	Status string `dynamodbav:"status" json:"status"`
}

// PaymentRecord abstrai o transacional legado do banco relacional (Postgres)
type PaymentRecord struct {
	ID          int     `json:"payment_id"`
	CPF         string  `json:"cpf"`
	Amount      float64 `json:"amount"`
	PaymentDate string  `json:"payment_date"`
	Status      string  `json:"status"`
}

// TransactionContext é a abstração unificada que entregamos ao modelo de IA
type TransactionContext struct {
	AnalysisTarget struct {
		CPF    string `json:"cpf"`
		Amount string `json:"amount"`
		Date   string `json:"date"`
	} `json:"analysis_target"`

	BusinessRules []Rule              `json:"business_rules_dynamo"`
	Liquidations  []LiquidationRecord `json:"liquidations_dynamo"`
	Payments      []PaymentRecord     `json:"payments_postgres"`
	ExternalAPI   string              `json:"api_inconsistencies"`
}

// ==========================================
// Inicialização e Handler MCP
// ==========================================

func main() {
	ctx := context.Background()

	// Inicializa Postgres (Dados de Pagamentos)
	pgDB, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/bank_data?sslmode=disable")
	if err != nil {
		log.Fatalf("Erro Postgres: %v", err)
	}
	defer pgDB.Close()

	// Inicializa DynamoDB (Regras e Dados de Baixa)
	cfg, _ := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	dynamoClient := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String("http://localhost:8000")
	})

	app := &App{DynamoClient: dynamoClient, PostgresDB: pgDB}

	s := server.NewMCPServer("Distributed-Consignado-MCP", "1.0.0")

	tool := mcp.NewTool("analyze_distributed_liquidation",
		mcp.WithDescription("Agrega regras (Dynamo) e dados de pagamentos/baixas (Postgres, Dynamo e APIs) para explicar inconsistências de crédito."),
		mcp.WithString("cpf", mcp.Required(), mcp.Description("CPF do cliente")),
		mcp.WithString("amount", mcp.Required(), mcp.Description("Valor da parcela")),
		mcp.WithString("date", mcp.Required(), mcp.Description("Data no formato YYYY-MM-DD")),
	)

	s.AddTool(tool, app.analyzeHandler)

	log.Println("MCP Server Híbrido rodando via STDIO...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Erro no servidor: %v", err)
	}
}

// ==========================================
// Implementações de Acesso a Dados
// ==========================================

func (a *App) fetchRulesFromDynamo(ctx context.Context) []Rule {
	params := &dynamodb.QueryInput{
		TableName:                aws.String("business_rules"),
		KeyConditionExpression:   aws.String("#d = :domainVal"),
		ExpressionAttributeNames: map[string]string{"#d": "domain"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":domainVal": &types.AttributeValueMemberS{Value: "baixa"},
		},
		ScanIndexForward: aws.Bool(true), // Garante ordenação pela Sort Key (execution_order)
	}

	out, err := a.DynamoClient.Query(ctx, params)
	if err != nil {
		return nil
	}
	var rules []Rule
	_ = attributevalue.UnmarshalListOfMaps(out.Items, &rules)
	return rules
}

func (a *App) fetchLiquidationsFromDynamo(ctx context.Context, cpf string) []LiquidationRecord {
	params := &dynamodb.QueryInput{
		TableName:              aws.String("liquidations"),
		KeyConditionExpression: aws.String("cpf = :c"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":c": &types.AttributeValueMemberS{Value: cpf},
		},
	}
	out, err := a.DynamoClient.Query(ctx, params)
	if err != nil {
		return nil
	}
	var records []LiquidationRecord
	_ = attributevalue.UnmarshalListOfMaps(out.Items, &records)
	return records
}

func (a *App) fetchPaymentsFromPostgres(cpf string) []PaymentRecord {
	query := `SELECT id, cpf, amount, payment_date, status FROM payments WHERE cpf = $1`
	rows, err := a.PostgresDB.Query(query, cpf)
	if err != nil || rows == nil {
		return nil
	}
	defer rows.Close()

	var records []PaymentRecord
	for rows.Next() {
		var p PaymentRecord
		var t time.Time
		if err := rows.Scan(&p.ID, &p.CPF, &p.Amount, &t, &p.Status); err == nil {
			p.PaymentDate = t.Format("2006-01-02")
			records = append(records, p)
		}
	}
	return records
}

// analyzeHandler orquestra a coleta concorrente
func (a *App) analyzeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Argumentos inválidos"), nil
	}

	cpf, _ := args["cpf"].(string)
	amount, _ := args["amount"].(string)
	date, _ := args["date"].(string)

	txCtx := TransactionContext{}
	txCtx.AnalysisTarget.CPF = cpf
	txCtx.AnalysisTarget.Amount = amount
	txCtx.AnalysisTarget.Date = date

	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(4) // 4 Fontes de Dados Distintas

	// 1. Fonte: DynamoDB (Regras de Negócio)
	go func() {
		defer wg.Done()
		rules := a.fetchRulesFromDynamo(ctx)
		mu.Lock()
		txCtx.BusinessRules = rules
		mu.Unlock()
	}()

	// 2. Fonte: DynamoDB (Dados de Baixa)
	go func() {
		defer wg.Done()
		liqs := a.fetchLiquidationsFromDynamo(ctx, cpf)
		mu.Lock()
		txCtx.Liquidations = liqs
		mu.Unlock()
	}()

	// 3. Fonte: PostgreSQL (Dados Transacionais/Pagamentos)
	go func() {
		defer wg.Done()
		payments := a.fetchPaymentsFromPostgres(cpf)
		mu.Lock()
		txCtx.Payments = payments
		mu.Unlock()
	}()

	// 4. Fonte: REST API (Mock de sistema externo)
	go func() {
		defer wg.Done()
		apiRes := mockRestAPIInconsistencies(cpf)
		mu.Lock()
		txCtx.ExternalAPI = apiRes
		mu.Unlock()
	}()

	wg.Wait()

	resultJSON, _ := json.MarshalIndent(txCtx, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

func mockRestAPIInconsistencies(cpf string) string {
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(40))) // Latência de rede
	return fmt.Sprintf("API de Custódia: Nenhuma divergência ativa no motor de crédito para o CPF %s.", cpf)
}
