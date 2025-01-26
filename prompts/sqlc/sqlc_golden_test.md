<anthropic_sqlc_golden_test_protocol>

## Core Architecture

### Test Suite Setup
```go
type myTestSuite struct {
    *testsuite.WPgxTestSuite
    usecase   *Usecase
    redisConn redis.UniversalClient
    cache     *dcache.DCache
}

func newMyTestSuite() *myTestSuite {
    return &myTestSuite{
        WPgxTestSuite: testsuite.NewWPgxTestSuiteFromEnv("testdb", []string{
            books.Schema,
            activities.Schema,
        }),
    }
}

func (s *myTestSuite) SetupTest() {
    s.WPgxTestSuite.SetupTest()
    s.cache.Clear()
    s.usecase = NewUsecase(s.GetPool(), s.cache)
}
```

### Serde Implementation
```go
type booksTableSerde struct {
    books *books.Queries
}

func (b booksTableSerde) Load(data []byte) error {
    err := b.books.Load(context.Background(), data)
    if err != nil {
        return err
    }
    return b.books.RefreshIDSerial(context.Background())
}

func (b booksTableSerde) Dump() ([]byte, error) {
    return b.books.Dump(context.Background(), func(m *books.Book) {
        m.CreatedAt = time.Unix(0, 0).UTC()
        m.UpdatedAt = time.Unix(0, 0).UTC()
    })
}
```

## Core Features

### State Management
- Load initial state:
```go
suite.LoadState("test_data.json", bookserde)
```

- Verify state after operations:
```go
suite.Golden("books_table", bookserde)
suite.Golden("activities_table", activitiesSerde)
```

### Time-Sensitive Testing
```go
// Template file: orders.json.tmpl
[
    {
        "id": 1,
        "created_at": "{{.P12H}}"  // 12 hours ago
    }
]

// Load with time points
timePoints := &TimePointSet{
    Now:  now,
    P12H: now.Add(-12 * time.Hour).Format(time.RFC3339),
}
suite.LoadStateTmpl("orders.json.tmpl", serde, timePoints)
```

### Variable Comparison
```go
// Compare complex structures
suite.GoldenVarJSON("result_name", resultVar)
```

## Best Practices

### State Handling
- Reset database state in `SetupTest()`
- Clear cache between tests
- Use `RefreshIDSerial` for auto-increment columns
- Normalize timestamps for comparisons

### Time Management
- Use templates for time-sensitive tests
- Normalize timestamps in Dump()
- Use relative time points in templates

### Cache Verification
- Clear cache in SetupTest
- Verify cache invalidation in transactions
- Test cache miss scenarios

### Transaction Testing
- Test complete transaction flows
- Verify state after rollbacks
- Check cache invalidation in transactions

## Key Points
1. Always call `SetupTest()` in subtests
2. Normalize volatile fields (timestamps, IDs)
3. Use templates for time-dependent tests
4. Clear caches between tests
5. Verify both success and failure cases
6. Test cache invalidation
7. Use GoldenVarJSON for complex comparisons

> Claude must follow this protocol in implementing golden tests.

</anthropic_sqlc_golden_test_protocol>