<anthropic_go_log_protocol>

## Core Principles

### Structured JSON Logging
- Always use structured logging with zerolog
- Log in JSON format for machine parsing
- Include consistent fields across all log entries
- Use strongly typed fields instead of string formatting

### Context Propagation
- Pass logger through context.Context
- Maintain trace ID across service boundaries
- Propagate context properly in goroutines
- Add request-scoped fields via context

### Performance
- Use sampling for high-volume debug logs
- Avoid expensive operations in log statements
- Leverage zerolog's zero-allocation design
- Check log level before expensive operations

## Logger Setup

### Base Configuration
```go
// Initialize global logger
log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

// Set global log level
zerolog.SetGlobalLevel(zerolog.InfoLevel)

// Configure timestamp format
zerolog.TimeFieldFormat = time.RFC3339Nano
```

### HTTP Middleware
```go
func TraceMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Generate trace ID
            traceID := fmt.Sprintf("%s-%d", uuid.NewString(), time.Now().Unix())
            
            // Create child logger with trace ID
            subLogger := log.With().
                Str("trace_id", traceID).
                Logger()
            
            // Add logger to context
            ctx := subLogger.WithContext(r.Context())
            
            // Add trace ID to response headers
            w.Header().Set("X-Trace-ID", traceID)
            
            // Serve with new context
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### gRPC Interceptor
```go
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    // Extract or generate trace ID
    traceID := GetTraceIDFromMetadata(ctx)
    if traceID == "" {
        traceID = uuid.NewString()
    }

    // Create scoped logger
    logger := log.With().
        Str("trace_id", traceID).
        Str("grpc_method", info.FullMethod).
        Logger()

    // Add to context
    newCtx := logger.WithContext(ctx)

    // Log request
    logger.Info().
        Interface("request", req).
        Msg("gRPC request received")

    // Handle with timing
    start := time.Now()
    resp, err := handler(newCtx, req)
    duration := time.Since(start)

    // Log result
    if err != nil {
        logger.Error().
            Err(err).
            Dur("duration_ms", duration).
            Msg("gRPC request failed")
    } else {
        logger.Info().
            Dur("duration_ms", duration).
            Msg("gRPC request succeeded")
    }

    return resp, err
}
```

## Usage Patterns

### Request Handler Logging
```go
func (h *Handler) HandleRequest(ctx context.Context, req Request) error {
    // Get logger from context
    logger := log.Ctx(ctx)

    logger.Info().
        Interface("request", req).
        Msg("Processing request")

    if err := h.process(ctx, req); err != nil {
        logger.Error().
            Err(err).
            Msg("Failed to process request")
        return err
    }

    logger.Info().Msg("Request processed successfully")
    return nil
}
```

### Goroutine Logging
```go
func ProcessAsync(ctx context.Context, data interface{}) {
    // Create child context with timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // Get parent logger
    parentLogger := log.Ctx(ctx)

    go func() {
        // Create goroutine-specific logger
        logger := parentLogger.With().
            Str("goroutine_id", fmt.Sprintf("%d", goid.Get())).
            Logger()
        
        ctx := logger.WithContext(ctx)

        logger.Debug().
            Interface("data", data).
            Msg("Starting async processing")

        // Process with context
        if err := process(ctx, data); err != nil {
            logger.Error().
                Err(err).
                Msg("Async processing failed")
            return
        }

        logger.Info().Msg("Async processing completed")
    }()
}
```

### Error Handling
```go
// Wrap errors with context
if err != nil {
    log.Ctx(ctx).Error().
        Err(err).
        Str("user_id", userID).
        Str("action", "create_user").
        Msg("Failed to create user")
    return fmt.Errorf("create user: %w", err)
}

// Log and return domain errors
if err == ErrUserNotFound {
    log.Ctx(ctx).Info().
        Str("user_id", userID).
        Msg("User not found")
    return err
}
```

## Best Practices

### Field Naming
- Use snake_case for field names
- Prefix service-specific fields
- Be consistent across services
- Use standard field names for common concepts

```go
logger.Info().
    Str("user_id", "123").
    Str("service_name", "auth").
    Str("http_method", "POST").
    Str("request_path", "/api/v1/users").
    Msg("Request received")
```

### Level Usage
- ERROR: Unexpected errors requiring attention
- WARN: Expected errors or important events
- INFO: Normal operations, state changes
- DEBUG: Detailed information for debugging
- TRACE: Very detailed debugging information

### Sampling Strategy
```go
// Sample debug logs
if log.Debug().Enabled() {
    log.Debug().
        Interface("payload", largePayload).
        Msg("Processing large payload")
}

// Configure sampling
logger = logger.Sample(&zerolog.BasicSampler{N: 10})
```

### Context Propagation
```go
// Add fields to context logger
ctx = log.Ctx(ctx).With().
    Str("request_id", requestID).
    Str("user_id", userID).
    Logger().WithContext(ctx)

// Use context logger
log.Ctx(ctx).Info().Msg("Processing request")
```

## Observability

### Log Querying
```bash
# Search by trace ID
{app="myapp"} |= "trace_id=abc-123"

# Search for errors in time range
{app="myapp"} |= "level=error" 
    | json 
    | duration > 1s 
    | time > 1h

# Aggregate by field
{app="myapp"} 
    | json 
    | count_over_time({http_status}[5m])
```

### Metrics Integration
```go
// Log with metrics
func trackRequest(ctx context.Context, start time.Time, method string, err error) {
    duration := time.Since(start)
    
    logger := log.Ctx(ctx).With().
        Str("method", method).
        Dur("duration_ms", duration)
    
    if err != nil {
        logger.Err(err).Msg("Request failed")
        metrics.RequestErrors.Inc()
    } else {
        logger.Info().Msg("Request succeeded")
        metrics.RequestSuccess.Inc()
    }
    
    metrics.RequestDuration.Observe(duration.Seconds())
}
```

### Trace Correlation
```go
// Add trace context
logger := log.With().
    Str("trace_id", span.TraceID()).
    Str("span_id", span.SpanID()).
    Logger()

// Log with trace context
logger.Info().
    Str("operation", op).
    Msg("Operation started")
```

## Common Patterns

### Request Context
```go
type RequestContext struct {
    TraceID    string
    RequestID  string
    UserID     string
    SessionID  string
}

func LogMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        
        reqCtx := RequestContext{
            TraceID:    uuid.NewString(),
            RequestID:  uuid.NewString(),
            UserID:     GetUserID(r),
            SessionID:  GetSessionID(r),
        }
        
        logger := log.With().
            Str("trace_id", reqCtx.TraceID).
            Str("request_id", reqCtx.RequestID).
            Str("user_id", reqCtx.UserID).
            Str("session_id", reqCtx.SessionID).
            Logger()
            
        ctx = context.WithValue(ctx, ctxKey, reqCtx)
        ctx = logger.WithContext(ctx)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Batch Processing
```go
func ProcessBatch(ctx context.Context, items []Item) error {
    logger := log.Ctx(ctx).With().
        Int("batch_size", len(items)).
        Logger()
    
    logger.Info().Msg("Starting batch processing")
    
    for i, item := range items {
        itemLogger := logger.With().
            Int("item_index", i).
            Interface("item_id", item.ID).
            Logger()
            
        if err := process(itemLogger.WithContext(ctx), item); err != nil {
            itemLogger.Error().
                Err(err).
                Msg("Failed to process item")
            continue
        }
        
        itemLogger.Debug().Msg("Processed item successfully")
    }
    
    logger.Info().Msg("Completed batch processing")
    return nil
}
```

### Background Tasks
```go
func StartBackgroundTask(ctx context.Context, task string) {
    logger := log.Ctx(ctx).With().
        Str("task", task).
        Str("worker_id", uuid.NewString()).
        Logger()
        
    go func() {
        ctx := logger.WithContext(ctx)
        
        ticker := time.NewTicker(time.Minute)
        defer ticker.Stop()
        
        for {
            select {
            case <-ctx.Done():
                logger.Info().Msg("Background task stopped")
                return
            case <-ticker.C:
                if err := runTask(ctx); err != nil {
                    logger.Error().
                        Err(err).
                        Msg("Background task failed")
                    continue
                }
                logger.Debug().Msg("Background task completed")
            }
        }
    }()
}
```

> Claude must follow this protocol when implementing logging in Go services.

</anthropic_go_log_protocol>
