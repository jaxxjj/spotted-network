<anthropic_go_concurrency_protocol>

## Core Principles

### Share Memory By Communicating
- Use channels for coordination between goroutines
- Avoid shared memory and mutexes when possible
- Think about communication patterns first

### Goroutines
- Lightweight execution units managed by Go runtime
- Start with `go` keyword
- Share address space but execute independently

```go
// Basic goroutine example
go func() {
    // Concurrent work here
}()
```

### Channels
- Primary mechanism for communication between goroutines
- Can be buffered or unbuffered
- Support synchronization and data transfer

```go
// Unbuffered channel
ch := make(chan int)

// Buffered channel with size 10
bufCh := make(chan int, 10)

// Send and receive
ch <- x     // Send x
x := <-ch   // Receive into x
```

## Concurrency Patterns

### Producer/Consumer
```go
func producer(ch chan<- int) {
    for i := 0; i < 10; i++ {
        ch <- i
    }
    close(ch)
}

func consumer(ch <-chan int) {
    for x := range ch {
        // Process x
    }
}
```

### Pipeline
```go
func stage1(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for x := range in {
            out <- x * 2
        }
    }()
    return out
}
```

### Fan-Out/Fan-In
```go
func fanOut(ch <-chan Work, n int) []<-chan Work {
    workers := make([]<-chan Work, n)
    for i := 0; i < n; i++ {
        workers[i] = worker(ch)
    }
    return workers
}

func fanIn(cs []<-chan Work) <-chan Work {
    var wg sync.WaitGroup
    out := make(chan Work)
    
    output := func(c <-chan Work) {
        defer wg.Done()
        for n := range c {
            out <- n
        }
    }
    
    wg.Add(len(cs))
    for _, c := range cs {
        go output(c)
    }
    
    go func() {
        wg.Wait()
        close(out)
    }()
    
    return out
}
```

## Error Handling

### Panic Recovery
```go
func worker() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from %v", r)
        }
    }()
    // Work that may panic
}
```

### Error Propagation
- Use error channels alongside data channels
- Consider context cancellation for cleanup
- Implement graceful shutdown

## Best Practices

### Resource Management
1. Always close channels from sender side
2. Use defer for cleanup
3. Implement proper cancellation
4. Check for channel closure when receiving

### Performance
1. Size buffers appropriately
2. Avoid goroutine leaks
3. Consider connection pooling
4. Use sync.Pool for frequent allocations

### Synchronization
1. Prefer channels over mutexes
2. Use sync.WaitGroup for fan-out
3. Consider sync.Once for initialization
4. Use atomic operations for simple counters

### Common Pitfalls
1. Sending on closed channel
2. Forgetting to close channels
3. Goroutine leaks
4. Race conditions with shared memory
5. Deadlocks from circular dependencies

## Testing

### Race Detection
```go
// Run tests with race detector
go test -race ./...
```

### Concurrency Testing
1. Use channels for synchronization in tests
2. Test timeout patterns
3. Verify goroutine cleanup
4. Check for race conditions

## Advanced Patterns

### Rate Limiting
```go
func rateLimiter(requests <-chan Request, rate time.Duration) <-chan Request {
    out := make(chan Request)
    tick := time.Tick(rate)
    go func() {
        for req := range requests {
            <-tick  // Rate limit
            out <- req
        }
    }()
    return out
}
```

### Context Usage
```go
func worker(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Do work
        }
    }
}
```

### Graceful Shutdown
```go
func server(stop <-chan struct{}) {
    for {
        select {
        case <-stop:
            // Cleanup
            return
        default:
            // Serve
        }
    }
}
```

## Modern Concurrency Patterns

### Worker Pools
```go
func worker(id int, jobs <-chan int, results chan<- int) {
    for j := range jobs {
        // Simulate work
        time.Sleep(time.Second)
        results <- j * 2
    }
}

// Create worker pool
jobs := make(chan int, 100)
results := make(chan int, 100)

for w := 1; w <= 3; w++ {
    go worker(w, jobs, results)
}
```

### Rate Limiting
```go
func rateLimiter(requests <-chan Request) <-chan Request {
    limiter := time.Tick(200 * time.Millisecond)
    out := make(chan Request)
    
    go func() {
        for req := range requests {
            <-limiter // Rate limit
            out <- req
        }
    }()
    return out
}
```

### Circuit Breaker
```go
type CircuitBreaker struct {
    threshold  int
    failures   int
    lastFail   time.Time
    timeout    time.Duration
    mu         sync.RWMutex
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    if !cb.canExecute() {
        return fmt.Errorf("circuit open")
    }
    
    if err := fn(); err != nil {
        cb.recordFailure()
        return err
    }
    
    cb.reset()
    return nil
}
```

### Backpressure
```go
func handleBackpressure(in <-chan Request, maxOutstanding int) {
    semaphore := make(chan struct{}, maxOutstanding)
    
    for req := range in {
        semaphore <- struct{}{} // Block if too many outstanding
        go func(r Request) {
            defer func() { <-semaphore }()
            process(r)
        }(req)
    }
}
```

## Best Practices

### Testing Concurrent Code
1. Use race detector
```go
go test -race ./...
```

2. Test timeouts and cancellation
```go
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    
    done := make(chan struct{})
    go func() {
        // Do work
        close(done)
    }()
    
    select {
    case <-done:
        // Success
    case <-ctx.Done():
        t.Fatal("timeout")
    }
}
```

3. Test concurrent access patterns
```go
func TestConcurrentAccess(t *testing.T) {
    const n = 1000
    var wg sync.WaitGroup
    counter := 0
    var mu sync.Mutex
    
    for i := 0; i < n; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            mu.Lock()
            counter++
            mu.Unlock()
        }()
    }
    
    wg.Wait()
    if counter != n {
        t.Errorf("got %d, want %d", counter, n)
    }
}
```

### Performance Optimization
1. Size channels appropriately
2. Avoid goroutine leaks
3. Use buffered channels when appropriate
4. Profile with pprof
5. Monitor goroutine count

### Graceful Shutdown
```go
func main() {
    srv := &http.Server{
        Addr: ":8080",
    }
    
    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit
    
    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## Pipeline Patterns

### Pipeline Definition
A pipeline is a series of stages connected by channels, where each stage:
- Receives values from upstream via inbound channels
- Performs computation on the data
- Sends values downstream via outbound channels

### Pipeline Stages
```go
// Source/Producer stage
func producer(done <-chan struct{}) <-chan int {
    stream := make(chan int)
    go func() {
        defer close(stream)
        for i := 0; i < 10; i++ {
            select {
            case <-done:
                return
            case stream <- i:
            }
        }
    }()
    return stream
}

// Processing stage
func processor(done <-chan struct{}, in <-chan int) <-chan int {
    stream := make(chan int)
    go func() {
        defer close(stream)
        for i := range in {
            select {
            case <-done:
                return
            case stream <- i * 2:
            }
        }
    }()
    return stream
}

// Sink/Consumer stage
func consumer(done <-chan struct{}, in <-chan int) {
    for val := range in {
        select {
        case <-done:
            return
        default:
            log.Printf("Consumed: %d", val)
        }
    }
}
```

### Pipeline Error Handling
```go
type Result struct {
    Value int
    Err   error
}

func errorAwarePipeline(done <-chan struct{}) <-chan Result {
    stream := make(chan Result)
    go func() {
        defer close(stream)
        for i := 0; i < 10; i++ {
            select {
            case <-done:
                return
            case stream <- Result{
                Value: i,
                Err:   validateValue(i),
            }:
            }
        }
    }()
    return stream
}

// Error propagation
func processWithErrors(in <-chan Result) <-chan Result {
    out := make(chan Result)
    go func() {
        defer close(out)
        for r := range in {
            if r.Err != nil {
                out <- r // Propagate error
                continue
            }
            // Process valid values
            value, err := process(r.Value)
            out <- Result{Value: value, Err: err}
        }
    }()
    return out
}
```

### Bounded Pipeline with Worker Pool
```go
func boundedWorkerPool(
    done <-chan struct{},
    inputs []int,
    numWorkers int,
) <-chan Result {
    // Create input and output channels
    inputChan := make(chan int)
    outputChan := make(chan Result)
    
    // Start fixed number of workers
    var wg sync.WaitGroup
    wg.Add(numWorkers)
    for i := 0; i < numWorkers; i++ {
        go func() {
            defer wg.Done()
            for input := range inputChan {
                select {
                case <-done:
                    return
                case outputChan <- process(input):
                }
            }
        }()
    }
    
    // Feed inputs
    go func() {
        defer close(inputChan)
        for _, input := range inputs {
            select {
            case <-done:
                return
            case inputChan <- input:
            }
        }
    }()
    
    // Close output when all workers done
    go func() {
        wg.Wait()
        close(outputChan)
    }()
    
    return outputChan
}
```

### Real-world Example: File Processing Pipeline
```go
// Process files in directory with bounded concurrency
func ProcessFiles(root string, numWorkers int) error {
    done := make(chan struct{})
    defer close(done)
    
    // Stage 1: Walk directory
    paths, errc := walkFiles(done, root)
    
    // Stage 2: Read and process files with worker pool
    results := make(chan Result)
    var wg sync.WaitGroup
    wg.Add(numWorkers)
    
    for i := 0; i < numWorkers; i++ {
        go func() {
            defer wg.Done()
            for path := range paths {
                data, err := ioutil.ReadFile(path)
                if err != nil {
                    select {
                    case results <- Result{Err: err}:
                    case <-done:
                        return
                    }
                    continue
                }
                
                // Process file data
                res, err := processData(data)
                select {
                case results <- Result{Value: res, Err: err}:
                case <-done:
                    return
                }
            }
        }()
    }
    
    // Close results when all workers done
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Stage 3: Handle results
    for r := range results {
        if r.Err != nil {
            return r.Err
        }
        // Handle successful result
    }
    
    // Check for walk error
    if err := <-errc; err != nil {
        return err
    }
    
    return nil
}

func walkFiles(done <-chan struct{}, root string) (<-chan string, <-chan error) {
    paths := make(chan string)
    errc := make(chan error, 1)
    
    go func() {
        defer close(paths)
        errc <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
            if err != nil {
                return err
            }
            if !info.Mode().IsRegular() {
                return nil
            }
            select {
            case paths <- path:
            case <-done:
                return errors.New("walk canceled")
            }
            return nil
        })
    }()
    
    return paths, errc
}
```

### Pipeline Best Practices

1. Error Handling
- Use Result type to wrap values and errors
- Propagate errors downstream
- Cancel pipeline on first error if appropriate
- Consider retry strategies for recoverable errors

2. Cancellation
- Pass done channel to all stages
- Check done channel in select statements
- Clean up resources on cancellation
- Use timeouts when appropriate

3. Performance
- Use bounded concurrency with worker pools
- Size buffers appropriately
- Monitor and handle backpressure
- Profile and optimize bottlenecks

4. Resource Management
- Close channels properly
- Clean up goroutines
- Handle panics
- Release file handles and other resources

5. Testing
- Test each stage independently
- Test error conditions
- Test cancellation
- Test with various concurrency levels

> Claude must follow these patterns when implementing concurrent Go programs.

</anthropic_go_concurrency_protocol>