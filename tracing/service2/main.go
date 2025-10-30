package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
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
	logger     *slog.Logger
}

func initLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
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

	rs.logger.InfoContext(ctx, "Processing order created event",
		slog.String("message_id", msg.UUID),
		slog.String("event_type", "OrderCreated"),
	)

	span.SetAttributes(
		attribute.String("message.id", msg.UUID),
		attribute.String("event.type", "OrderCreated"),
	)

	var event OrderCreatedEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		rs.logger.ErrorContext(ctx, "Failed to unmarshal order created event",
			slog.String("error", err.Error()),
			slog.String("message_id", msg.UUID),
		)
		return err
	}

	rs.logger.InfoContext(ctx, "Order event unmarshaled successfully",
		slog.String("order_id", event.OrderID),
		slog.Int("total_price", event.TotalPrice),
		slog.Int("customer_id", event.CustomerID),
		slog.Int("product_id", event.ProductID),
	)

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
		rs.logger.WarnContext(ctx, "Failed to parse created_at time, using current time",
			slog.String("error", err.Error()),
			slog.String("order_id", event.OrderID),
			slog.String("created_at_raw", event.CreatedAt),
		)
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
	reportCount := len(rs.reports)
	rs.mu.Unlock()

	rs.logger.InfoContext(ctx, "Order report processed and stored",
		slog.String("order_id", report.OrderID),
		slog.Int("total_price", report.TotalPrice),
		slog.Int("customer_id", report.CustomerID),
		slog.Int("product_id", report.ProductID),
		slog.Time("created_at", report.CreatedAt),
		slog.Time("processed_at", report.ProcessedAt),
		slog.Int("total_reports", reportCount),
	)

	// Acknowledge the message
	msg.Ack()
	return nil
}

func (rs *ReportService) getReports(c *gin.Context) {
	rs.mu.RLock()
	reports := make([]OrderReport, len(rs.reports))
	copy(reports, rs.reports)
	reportCount := len(reports)
	rs.mu.RUnlock()

	rs.logger.Info("Reports retrieved",
		slog.Int("total_reports", reportCount),
		slog.String("request_id", c.GetHeader("X-Request-ID")),
	)

	c.JSON(http.StatusOK, gin.H{
		"reports": reports,
		"total":   reportCount,
	})
}

func (rs *ReportService) startMessageConsumer(ctx context.Context) error {
	rs.logger.Info("Starting message consumer", slog.String("topic", "orders"))

	messages, err := rs.subscriber.Subscribe(ctx, "orders")
	if err != nil {
		rs.logger.Error("Failed to subscribe to orders topic", slog.String("error", err.Error()))
		return err
	}

	go func() {
		rs.logger.Info("Message consumer started, listening for messages")
		for msg := range messages {
			if err := rs.handleOrderCreated(msg); err != nil {
				rs.logger.Error("Failed to handle message",
					slog.String("error", err.Error()),
					slog.String("message_id", msg.UUID),
				)
				msg.Nack()
			}
		}
		rs.logger.Warn("Message consumer stopped")
	}()

	return nil
}

func main() {
	// Initialize structured logger
	logger := initLogger()
	logger.Info("Starting report service", slog.String("version", "v1.0.0"))

	// Initialize tracing
	tp, err := initTracer()
	if err != nil {
		logger.Error("Failed to initialize tracer", slog.String("error", err.Error()))
		log.Fatal("Failed to initialize tracer:", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("Error shutting down tracer provider", slog.String("error", err.Error()))
		}
	}()
	logger.Info("Tracer initialized successfully")

	// Initialize Watermill subscriber
	subscriber, err := initWatermill()
	if err != nil {
		logger.Error("Failed to initialize Watermill", slog.String("error", err.Error()))
		log.Fatal("Failed to initialize Watermill:", err)
	}
	defer subscriber.Close()
	logger.Info("Watermill subscriber initialized successfully")

	// Initialize service
	reportService := &ReportService{
		subscriber: subscriber,
		tracer:     otel.Tracer("report-service"),
		reports:    make([]OrderReport, 0),
		logger:     logger,
	}

	// Start message consumer
	ctx := context.Background()
	if err := reportService.startMessageConsumer(ctx); err != nil {
		logger.Error("Failed to start message consumer", slog.String("error", err.Error()))
		log.Fatal("Failed to start message consumer:", err)
	}

	// Initialize Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("report-service"))

	// Add logging middleware
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("HTTP request completed",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", time.Since(start)),
			slog.String("client_ip", c.ClientIP()),
		)
	})

	// Routes
	r.GET("/reports", reportService.getReports)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	logger.Info("Report service starting", slog.String("port", ":8081"))
	if err := r.Run(":8081"); err != nil {
		logger.Error("Failed to start server", slog.String("error", err.Error()))
		log.Fatal("Failed to start server:", err)
	}
}
