package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

type OrderRequest struct {
	TotalPrice int `json:"total_price" binding:"required"`
	CustomerID int `json:"customer_id" binding:"required"`
	ProductID  int `json:"product_id" binding:"required"`
}

type OrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

type OrderCreatedEvent struct {
	OrderID    string `json:"order_id"`
	TotalPrice int    `json:"total_price"`
	CustomerID int    `json:"customer_id"`
	ProductID  int    `json:"product_id"`
	CreatedAt  string `json:"created_at"`
}

type OrderService struct {
	publisher message.Publisher
	tracer    trace.Tracer
	logger    *slog.Logger
}

func initLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

func initTracer() (*sdktrace.TracerProvider, error) {
	logger := slog.Default()
	logger.Info("Initializing tracer", "endpoint", "jaeger:4318")

	// Create OTLP HTTP exporter
	exp, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint("jaeger:4318"),
		otlptracehttp.WithURLPath("/v1/traces"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		logger.Error("Failed to create OTLP exporter", "error", err, "endpoint", "jaeger:4318")
		return nil, err
	}

	logger.Info("OTLP exporter created successfully")

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("order-service"),
			semconv.ServiceVersionKey.String("v1.0.0"),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	logger.Info("Tracer initialized successfully", "service_name", "order-service", "version", "v1.0.0")
	return tp, nil
}

func initWatermill() (message.Publisher, error) {
	logger := slog.Default()
	logger.Info("Initializing Watermill publisher", "rabbitmq_url", "amqp://guest:guest@rabbitmq:5672/")

	amqpConfig := amqp.NewDurableQueueConfig("amqp://guest:guest@rabbitmq:5672/")

	publisher, err := amqp.NewPublisher(amqpConfig, watermill.NewStdLogger(false, false))
	if err != nil {
		logger.Error("Failed to create Watermill publisher", "error", err, "rabbitmq_url", "amqp://guest:guest@rabbitmq:5672/")
		return nil, err
	}

	logger.Info("Watermill publisher initialized successfully", "publisher_type", "amqp")
	return publisher, nil
}

func (os *OrderService) createOrder(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)

	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	os.logger.Info("Received create order request",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"trace_id", traceID,
		"span_id", spanID,
	)

	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		os.logger.Error("Invalid order request - JSON binding failed",
			"error", err,
			"content_type", c.GetHeader("Content-Type"),
			"trace_id", traceID,
			"span_id", spanID,
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	os.logger.Info("Order request parsed successfully",
		"total_price", req.TotalPrice,
		"customer_id", req.CustomerID,
		"product_id", req.ProductID,
		"trace_id", traceID,
		"span_id", spanID,
	)

	// Create order ID
	orderID := uuid.New().String()

	os.logger.Info("Generated order ID",
		"order_id", orderID,
		"trace_id", traceID,
		"span_id", spanID,
	)

	// Create order created event
	event := OrderCreatedEvent{
		OrderID:    orderID,
		TotalPrice: req.TotalPrice,
		CustomerID: req.CustomerID,
		ProductID:  req.ProductID,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	os.logger.Info("Order event created",
		"order_id", orderID,
		"event_type", "OrderCreated",
		"created_at", event.CreatedAt,
		"trace_id", traceID,
		"span_id", spanID,
	)

	// Publish event
	if err := os.publishOrderCreatedEvent(ctx, event); err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		os.logger.Error("Failed to publish order created event",
			"error", err,
			"order_id", orderID,
			"event_type", "OrderCreated",
			"trace_id", traceID,
			"span_id", spanID,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish event"})
		return
	}

	response := OrderResponse{
		OrderID: orderID,
		Status:  "created",
	}

	os.logger.Info("Order created successfully",
		"order_id", orderID,
		"status", response.Status,
		"http_status", http.StatusCreated,
		"trace_id", traceID,
		"span_id", spanID,
	)

	c.JSON(http.StatusCreated, response)
}

func (os *OrderService) publishOrderCreatedEvent(ctx context.Context, event OrderCreatedEvent) error {
	// Create a child span for publishing
	ctx, span := os.tracer.Start(ctx, "publish_order_created_event")
	defer span.End()

	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	os.logger.Info("Starting event publication",
		"operation", "publish_order_created_event",
		"order_id", event.OrderID,
		"trace_id", traceID,
		"span_id", spanID,
	)

	eventData, err := json.Marshal(event)
	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		os.logger.Error("Failed to marshal event data to JSON",
			"error", err,
			"order_id", event.OrderID,
			"event_type", "OrderCreated",
			"trace_id", traceID,
			"span_id", spanID,
		)
		return err
	}

	os.logger.Info("Event data marshaled successfully",
		"order_id", event.OrderID,
		"data_size", len(eventData),
		"trace_id", traceID,
		"span_id", spanID,
	)

	// Create message with trace context
	msg := message.NewMessage(watermill.NewUUID(), eventData)

	os.logger.Info("Message created",
		"message_id", msg.UUID,
		"order_id", event.OrderID,
		"trace_id", traceID,
		"span_id", spanID,
	)

	// Inject trace context into message headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(msg.Metadata))

	os.logger.Info("Trace context injected into message headers",
		"message_id", msg.UUID,
		"metadata_keys", len(msg.Metadata),
		"trace_id", traceID,
		"span_id", spanID,
	)

	span.SetAttributes(
		attribute.String("message.id", msg.UUID),
		attribute.String("exchange", "orders"),
		attribute.String("event.type", "OrderCreated"),
	)

	os.logger.Info("Publishing message to exchange",
		"message_id", msg.UUID,
		"exchange", "orders",
		"event_type", "OrderCreated",
		"order_id", event.OrderID,
		"trace_id", traceID,
		"span_id", spanID,
	)

	err = os.publisher.Publish("orders", msg)
	if err != nil {
		os.logger.Error("Failed to publish message to exchange",
			"error", err,
			"message_id", msg.UUID,
			"exchange", "orders",
			"order_id", event.OrderID,
			"trace_id", traceID,
			"span_id", spanID,
		)
		return err
	}

	os.logger.Info("Message published successfully to exchange",
		"message_id", msg.UUID,
		"exchange", "orders",
		"order_id", event.OrderID,
		"trace_id", traceID,
		"span_id", spanID,
	)

	return nil
}

func main() {
	// Initialize structured logger
	logger := initLogger()
	slog.SetDefault(logger)

	logger.Info("Order service initializing", "version", "v1.0.0")

	// Initialize tracing
	tp, err := initTracer()
	if err != nil {
		logger.Error("Failed to initialize tracer", "error", err)
		log.Fatal("Failed to initialize tracer:", err)
	}
	defer func() {
		logger.Info("Shutting down tracer provider")
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("Error shutting down tracer provider", "error", err)
		} else {
			logger.Info("Tracer provider shutdown completed")
		}
	}()

	// Initialize Watermill publisher
	publisher, err := initWatermill()
	if err != nil {
		logger.Error("Failed to initialize Watermill", "error", err)
		log.Fatal("Failed to initialize Watermill:", err)
	}
	defer func() {
		logger.Info("Closing Watermill publisher")
		publisher.Close()
		logger.Info("Watermill publisher closed")
	}()

	// Initialize service
	orderService := &OrderService{
		publisher: publisher,
		tracer:    otel.Tracer("order-service"),
		logger:    logger,
	}

	logger.Info("Order service initialized", "tracer_name", "order-service")

	// Initialize Gin
	gin.SetMode(gin.ReleaseMode)
	logger.Info("Gin mode set", "mode", "release")

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("order-service"))

	logger.Info("Gin middleware configured", "middlewares", []string{"Recovery", "OpenTelemetry"})

	// Routes
	r.POST("/order", orderService.createOrder)
	r.GET("/health", func(c *gin.Context) {
		logger.Info("Health check requested",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"user_agent", c.GetHeader("User-Agent"),
		)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	logger.Info("Routes registered", "routes", []string{"POST /order", "GET /health"})

	logger.Info("Order service starting", "port", "8080", "protocol", "http")
	if err := r.Run(":8080"); err != nil {
		logger.Error("Failed to start HTTP server", "error", err, "port", "8080")
		log.Fatal("Failed to start server:", err)
	}
}
