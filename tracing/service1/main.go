package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
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
}

func initTracer() (*sdktrace.TracerProvider, error) {
	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://jaeger:14268/api/traces")))
	if err != nil {
		return nil, err
	}

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
	return tp, nil
}

func initWatermill() (message.Publisher, error) {
	amqpConfig := amqp.NewDurableQueueConfig("amqp://guest:guest@rabbitmq:5672/")

	publisher, err := amqp.NewPublisher(amqpConfig, watermill.NewStdLogger(false, false))
	if err != nil {
		return nil, err
	}

	return publisher, nil
}

func (os *OrderService) createOrder(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)

	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create order ID
	orderID := uuid.New().String()

	span.SetAttributes(
		attribute.String("order.id", orderID),
		attribute.Int("order.total_price", req.TotalPrice),
		attribute.Int("order.customer_id", req.CustomerID),
		attribute.Int("order.product_id", req.ProductID),
	)

	// Create order created event
	event := OrderCreatedEvent{
		OrderID:    orderID,
		TotalPrice: req.TotalPrice,
		CustomerID: req.CustomerID,
		ProductID:  req.ProductID,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	// Publish event
	if err := os.publishOrderCreatedEvent(ctx, event); err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		log.Printf("Failed to publish order created event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish event"})
		return
	}

	response := OrderResponse{
		OrderID: orderID,
		Status:  "created",
	}

	c.JSON(http.StatusCreated, response)
}

func (os *OrderService) publishOrderCreatedEvent(ctx context.Context, event OrderCreatedEvent) error {
	// Create a child span for publishing
	ctx, span := os.tracer.Start(ctx, "publish_order_created_event")
	defer span.End()

	eventData, err := json.Marshal(event)
	if err != nil {
		span.SetAttributes(attribute.String("error", err.Error()))
		return err
	}

	// Create message with trace context
	msg := message.NewMessage(watermill.NewUUID(), eventData)

	// Inject trace context into message headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(msg.Metadata))

	span.SetAttributes(
		attribute.String("message.id", msg.UUID),
		attribute.String("exchange", "orders"),
		attribute.String("event.type", "OrderCreated"),
	)

	return os.publisher.Publish("orders", msg)
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

	// Initialize Watermill publisher
	publisher, err := initWatermill()
	if err != nil {
		log.Fatal("Failed to initialize Watermill:", err)
	}
	defer publisher.Close()

	// Initialize service
	orderService := &OrderService{
		publisher: publisher,
		tracer:    otel.Tracer("order-service"),
	}

	// Initialize Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("order-service"))

	// Routes
	r.POST("/order", orderService.createOrder)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Println("Order service starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
