package main

import (
	"context"
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"gopkg.in/yaml.v3"
)

// Estrutura YAML mapeada diretamente para as tags do DynamoDB
type RuleDocument struct {
	RuleID         string `yaml:"rule_id" dynamodbav:"rule_id"`
	Domain         string `yaml:"domain" dynamodbav:"domain"`
	Context        string `yaml:"context" dynamodbav:"context"`
	ExecutionOrder int    `yaml:"execution_order" dynamodbav:"execution_order"`
	Status         string `yaml:"status" dynamodbav:"status"`
	AILogic        string `yaml:"ai_logic" dynamodbav:"ai_logic"`

	HumanContext struct {
		Name          string `yaml:"name" dynamodbav:"name"`
		Description   string `yaml:"description" dynamodbav:"description"`
		BusinessOwner string `yaml:"business_owner" dynamodbav:"business_owner"`
	} `yaml:"human_context" dynamodbav:"human_context"`

	EngineeringContext struct {
		ApplicationName string `yaml:"application_name" dynamodbav:"application_name"`
		ApplicationType string `yaml:"application_type" dynamodbav:"application_type"`
		RepositoryURL   string `yaml:"repository_url" dynamodbav:"repository_url"`
		Entrypoint      string `yaml:"entrypoint" dynamodbav:"entrypoint"`
	} `yaml:"engineering_context" dynamodbav:"engineering_context"`

	TechnicalMetadata struct {
		Dependencies []struct {
			System string `yaml:"system" dynamodbav:"system"`
			Action string `yaml:"action" dynamodbav:"action"`
		} `yaml:"dependencies" dynamodbav:"dependencies"`
		Observability struct {
			DatadogMonitorID string `yaml:"datadog_monitor_id" dynamodbav:"datadog_monitor_id"`
			CustomMetric     string `yaml:"custom_metric" dynamodbav:"custom_metric"`
			LogMarker        string `yaml:"log_marker" dynamodbav:"log_marker"`
		} `yaml:"observability" dynamodbav:"observability"`
	} `yaml:"technical_metadata" dynamodbav:"technical_metadata"`
}

func init() {
	mode := flag.String("mode", "server", "Modo de execução: 'server' ou 'import'")

	flag.Parse()

	if *mode != "server" && *mode != "import" {
		log.Fatalf("Modo inválido: %s. Use 'server' ou 'import'.", *mode)
	} else if *mode == "import" {
		importRules()
		os.Exit(0)
	}
}

func importRules() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("Erro aws config: %v", err)
	}

	dynamoClient := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String("http://localhost:8000")
	})

	// Caminho onde os YAMLs ficarão armazenados
	rulesDir := "./rules"

	err = filepath.WalkDir(rulesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".yaml" {
			return err
		}

		fileData, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Erro ao ler %s: %v", path, err)
			return nil
		}

		var rule RuleDocument
		if err := yaml.Unmarshal(fileData, &rule); err != nil {
			log.Printf("Erro de parse no YAML %s: %v", path, err)
			return nil
		}

		item, err := attributevalue.MarshalMap(rule)
		if err != nil {
			log.Printf("Erro ao fazer marshal pro DynamoDB %s: %v", path, err)
			return nil
		}

		_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String("business_rules"),
			Item:      item,
		})

		if err != nil {
			log.Printf("Falha ao salvar %s: %v", rule.RuleID, err)
		} else {
			log.Printf("Regra %s processada e salva com sucesso.", rule.RuleID)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Erro na varredura: %v", err)
	}
}
