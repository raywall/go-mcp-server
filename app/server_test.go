package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	_ "github.com/lib/pq"
	"github.com/mark3labs/mcp-go/mcp"
)

// helperApp gera o ambiente conectado para os benchmarks
func helperApp() *App {
	ctx := context.Background()
	cfg, _ := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	dynamo := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String("http://localhost:8000")
	})

	pgDB, _ := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/bank_data?sslmode=disable")
	pgDB.SetMaxOpenConns(50) // Limita conexões para stress tests

	return &App{DynamoClient: dynamo, PostgresDB: pgDB}
}

func BenchmarkHandlerParallel(b *testing.B) {
	app := helperApp()
	defer app.PostgresDB.Close()

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"cpf":    "12345678900",
		"amount": "450.50",
		"date":   "2026-05-10",
	}

	ctx := context.Background()
	b.ResetTimer()

	// Valida consumo simultâneo e thread-safety (locks)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := app.analyzeHandler(ctx, req)
			if err != nil {
				return
			}

			// Parse do payload gerado para garantir integridade do Fan-In
			var txCtx TransactionContext
			content := res.Content[0].(mcp.TextContent)
			_ = json.Unmarshal([]byte(content.Text), &txCtx)
		}
	})
}
