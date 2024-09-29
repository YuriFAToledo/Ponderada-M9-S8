package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
    "google.golang.org/grpc/credentials"
)

var (
    serviceName  = os.Getenv("SERVICE_NAME")
    collectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
    insecure     = os.Getenv("INSECURE_MODE")
)

// Função para inicializar o rastreador OpenTelemetry
func initTracer() func(context.Context) error {
    secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
    if len(insecure) > 0 {
        secureOption = otlptracegrpc.WithInsecure()
    }

    exporter, err := otlptrace.New(
        context.Background(),
        otlptracegrpc.NewClient(
            secureOption,
            otlptracegrpc.WithEndpoint(collectorURL),
        ),
    )
    if err != nil {
        log.Fatal(err)
    }

    resources, err := resource.New(
        context.Background(),
        resource.WithAttributes(
            attribute.String("service.name", serviceName),
            attribute.String("library.language", "go"),
        ),
    )
    if err != nil {
        log.Printf("Could not set resources: %v", err)
    }

    otel.SetTracerProvider(
        sdktrace.NewTracerProvider(
            sdktrace.WithSampler(sdktrace.AlwaysSample()),
            sdktrace.WithBatcher(exporter),
            sdktrace.WithResource(resources),
        ),
    )

    return exporter.Shutdown
}

func main() {
    // Inicializar o rastreador OpenTelemetry
    cleanup := initTracer()
    defer cleanup(context.Background())

    // Criar a instância do Gin e adicionar o middleware OpenTelemetry
    r := gin.Default()
    r.Use(otelgin.Middleware(serviceName))

    // Definir um endpoint simples
    r.GET("/books", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "Hello, OpenTelemetry with Gin!",
        })
    })

    // Rodar o servidor na porta 8090
    fmt.Println("Server running at http://localhost:8090")
    if err := r.Run(":8090"); err != nil {
        log.Fatal(err)
    }
}
