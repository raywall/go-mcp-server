package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// BenchmarkAnalyzeLiquidationSequential mede a performance do handler em um cenário linear (single-thread).
// Permite entender o custo de CPU, tempo de execução e alocações de memória por operação.
func BenchmarkAnalyzeLiquidationSequential(b *testing.B) {
	// Setup da requisição mockada imitando o payload do LLM
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"cpf":    "12345678900",
		"amount": "450.50",
		"date":   "2026-05-10",
	}

	ctx := context.Background()

	// Reseta o cronômetro para desconsiderar o tempo de alocação inicial do setup
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		res, err := analyzeLiquidationHandler(ctx, req)
		if err != nil {
			b.Fatalf("Erro inesperado no handler: %v", err)
		}

		// Validação mínima para garantir que o compilador não otimize o loop eliminando o código morto
		if res == nil {
			b.Fatal("Resultado retornado é nulo")
		}
	}
}

// BenchmarkAnalyzeLiquidationParallel mede a escalabilidade e o comportamento sob concorrência.
// Simula múltiplas threads do servidor MCP chamando a ferramenta simultaneamente (Stress Test).
func BenchmarkAnalyzeLiquidationParallel(b *testing.B) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"cpf":    "98765432100",
		"amount": "1200.00",
		"date":   "2026-05-12",
	}

	ctx := context.Background()
	b.ResetTimer()

	// Executa o benchmark utilizando a quantidade máxima de cores (GOMAXPROCS) configurada no sistema
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := analyzeLiquidationHandler(ctx, req)
			if err != nil {
				return
			}

			// Garante o parse do JSON interno para medir o impacto de marshaling sob concorrência
			var txCtx TransactionContext
			content := res.Content[0].(mcp.TextContent)
			_ = json.Unmarshal([]byte(content.Text), &txCtx)
		}
	})
}
