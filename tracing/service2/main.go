package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type OrderCreatedEvent struct {
	OrderID    string `json:"order_id"`
	TotalPrice int    `json:"total_price"`
	CustomerID int    `json:"customer_id"`
	ProductID  int    `json:"product_id"`
	CreatedAt  string `json:"created_at"`
}

type OrderReport struct {
	OrderID     string    `json:"order_id"`
	TotalPrice  int       `json:"total_price"`
	CustomerID  int       `json:"customer_id"`
	ProductID   int       `json:"product_id"`
	CreatedAt   time.Time `json:"created_at"`
	ProcessedAt time.Time `json:"processed_at"`
}

type ReportService struct {
	subscriber message.Subscriber
	tracer     trace.Tracer
	reports    []OrderReport
	mu         sync.RWMutex
}

func initTracer() (*sdktrace.TracerProvider, error) {
	// Create OTLP HTTP exporter
	exp, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint("jaeger:4318"),
		otlptracehttp.WithURLPath("/v1/traces"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("report-service"),
			semconv.ServiceVersionKey.String("v1.0.0"),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return tp, nil
}

func initWatermill() (message.Subscriber, error) {
	amqpConfig := amqp.NewDurableQueueConfig("amqp://guest:guest@rabbitmq:5672/")

	subscriber, err := amqp.NewSubscriber(amqpConfig, watermill.NewStdLogger(false, false))
	if err != nil {
		return nil, err
	}

	return subscriber, nil
}

func (rs *ReportService) handleOrderCreated(msg *message.Message) error {
	// Extract trace context from message headers
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(msg.Metadata))

	// Create a new span with the extracted context
	ctx, span := rs.tracer.Start(ctx, "process_order_created_event")
	defer span.End()

	span.SetAttributes(
		attribute.String("message.id", msg.UUID),
		attribute.String("event.type", "OrderCreated"),
	)

	var event OrderCreatedEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		log.Printf("Failed to unmarshal order created event: %v", err)
		return err
	}

	span.SetAttributes(
		attribute.String("order.id", event.OrderID),
		attribute.Int("order.total_price", event.TotalPrice),
		attribute.Int("order.customer_id", event.CustomerID),
		attribute.Int("order.product_id", event.ProductID),
	)

	// Parse created time
	createdAt, err := time.Parse(time.RFC3339, event.CreatedAt)
	if err != nil {
		span.SetAttributes(attribute.String("parse_error", err.Error()))
		log.Printf("Failed to parse created_at time: %v", err)
		createdAt = time.Now().UTC()
	}

	// Create order report
	report := OrderReport{
		OrderID:     event.OrderID,
		TotalPrice:  event.TotalPrice,
		CustomerID:  event.CustomerID,
		ProductID:   event.ProductID,
		CreatedAt:   createdAt,
		ProcessedAt: time.Now().UTC(),
	}

	// Store report (in memory for demo purposes)
	rs.mu.Lock()
	rs.reports = append(rs.reports, report)
	rs.mu.Unlock()

	log.Printf("Processed order report: OrderID=%s, TotalPrice=%d, CustomerID=%d, ProductID=%d",
		report.OrderID, report.TotalPrice, report.CustomerID, report.ProductID)

	// Acknowledge the message
	msg.Ack()
	return nil
}

func (rs *ReportService) getReports(c *gin.Context) {
	rs.mu.RLock()
	reports := make([]OrderReport, len(rs.reports))
	copy(reports, rs.reports)
	rs.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"reports": reports,
		"total":   len(reports),
	})
}

func (rs *ReportService) startMessageConsumer(ctx context.Context) error {
	messages, err := rs.subscriber.Subscribe(ctx, "orders")
	if err != nil {
		return err
	}

	go func() {
		for msg := range messages {
			if err := rs.handleOrderCreated(msg); err != nil {
				log.Printf("Failed to handle message: %v", err)
				msg.Nack()
			}
		}
	}()

	return nil
}

func main() {
	// Initialize tracing
	tp, err := initTracer()
	if err != nil {
		log.Fatal("Failed to initialize tracer:", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Initialize Watermill subscriber
	subscriber, err := initWatermill()
	if err != nil {
		log.Fatal("Failed to initialize Watermill:", err)
	}
	defer subscriber.Close()

	// Initialize service
	reportService := &ReportService{
		subscriber: subscriber,
		tracer:     otel.Tracer("report-service"),
		reports:    make([]OrderReport, 0),
	}

	// Start message consumer
	ctx := context.Background()
	if err := reportService.startMessageConsumer(ctx); err != nil {
		log.Fatal("Failed to start message consumer:", err)
	}

	// Initialize Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("report-service"))

	// Routes
	r.GET("/reports", reportService.getReports)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Println("Report service starting on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
