# Testing Protocol for Go Services

## Core Testing Patterns

### 1. Table Driven Tests

Table Driven Tests are part of Google's coding style and help standardize tests while reducing code duplication. For large projects, test code maintainability is as important as the code itself.

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected Result
        wantErr  error
    }{
        {
            name:     "normal case",
            input:    "test",
            expected: expectedResult,
            wantErr:  nil,
        },
        {
            name:     "error case",
            input:    "invalid",
            expected: Result{},
            wantErr:  ErrInvalid,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)
            if tt.wantErr != nil {
                assert.Equal(t, tt.wantErr, err)
                return
            }
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 2. Golden Tests

Golden tests are particularly useful for testing complex data structures and database states.

```go
type TableSerde struct {
    repo *Repo
}

func (s TableSerde) Load(data []byte) error {
    err := s.repo.Load(context.Background(), data)
    if err != nil {
        return err
    }
    return s.repo.RefreshIDSerial(context.Background())
}

func (s TableSerde) Dump() ([]byte, error) {
    return s.repo.Dump(context.Background(), func(m *Model) {
        // Normalize timestamps
        m.CreatedAt = time.Unix(0, 0).UTC()
        m.UpdatedAt = time.Unix(0, 0).UTC()
    })
}
```

### 3. Test Suite Setup

```go
type MyTestSuite struct {
    *testsuite.WPgxTestSuite
    usecase   *Usecase
    redisConn redis.UniversalClient
    cache     *dcache.DCache
}

func newMyTestSuite() *MyTestSuite {
    return &MyTestSuite{
        WPgxTestSuite: testsuite.NewWPgxTestSuiteFromEnv("testdb", []string{
            schema1,
            schema2,
        }),
    }
}

func TestMyTestSuite(t *testing.T) {
    suite.Run(t, newMyTestSuite())
}
```

## Testing Best Practices

### 1. State Management

```go
// Load initial state
suite.LoadState("test_data.json", serde)

// Verify final state
suite.Golden("table_state", serde)
suite.GoldenVarJSON("complex_result", result)
```

### 2. Database Testing

```go
func (suite *MyTestSuite) TestDatabaseOperation() {
    // Prepare data
    suite.LoadState("initial_state.json", serde)
    
    // Execute operation
    err := suite.usecase.Operation(ctx)
    suite.NoError(err)
    
    // Verify result
    suite.Golden("final_state", serde)
}
```

### 3. Cache Testing

```go
func (suite *MyTestSuite) TestCache() {
    // Clear cache
    suite.cache.Clear()
    
    // First query
    result1, err := suite.usecase.CachedQuery()
    suite.NoError(err)
    
    // Verify cache hit
    result2, err := suite.usecase.CachedQuery()
    suite.NoError(err)
    suite.Equal(result1, result2)
}
```

### 4. Parallel Testing

```go
func TestParallel(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"test1", "input1"},
        {"test2", "input2"},
    }
    
    for _, tt := range tests {
        tt := tt // Create local copy
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // Test logic
        })
    }
}
```

## Key Testing Principles

1. **Test Independence**
   - Each test should be independent
   - No dependencies between tests
   - Reset state before each test

2. **Test Readability**
   - Use clear test names
   - Provide detailed error messages
   - Use Table Driven Tests to organize cases

3. **Test Completeness**
   - Test normal paths
   - Test error paths
   - Test edge cases

4. **Test Efficiency**
   - Use parallel testing where appropriate
   - Optimize test data preparation
   - Use test caching wisely

5. **Test Maintainability**
   - Use Golden Tests for complex data
   - Extract common test logic
   - Keep test code clean

## Database Testing with WPgx

For services depending on PostgreSQL, use wpgx testsuite:

```go
import (
    "github.com/stumble/wpgx/testsuite"
)

type myTestSuite struct {
    *testsuite.WPgxTestSuite
}

func newMyTestSuite() *myTestSuite {
    return &myTestSuite{
        WPgxTestSuite: testsuite.NewWPgxTestSuiteFromEnv("testDbName", []string{
            `CREATE TABLE IF NOT EXISTS books (
                // Schema definition
            );`,
            // Add other schemas
        }),
    }
}

func TestMyTestSuite(t *testing.T) {
    suite.Run(t, newMyTestSuite())
}
```

## Time-Sensitive Testing

```go
// Template file: data.json.tmpl
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
suite.LoadStateTmpl("data.json.tmpl", serde, timePoints)
```

## Error Testing

```go
func TestErrors(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr error
    }{
        {
            name:    "invalid input",
            input:   "invalid",
            wantErr: ErrInvalidInput,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := Function(tt.input)
            assert.Equal(t, tt.wantErr, err)
        })
    }
}
```

## Best Practices Summary

1. Always call `SetupTest()` in subtests
2. Normalize volatile fields (timestamps, IDs)
3. Use templates for time-dependent tests
4. Clear caches between tests
5. Verify both success and failure cases
6. Test cache invalidation
7. Use GoldenVarJSON for complex comparisons
8. Implement proper cleanup in `TearDownTest()`
9. Use meaningful test names
10. Provide detailed error messages

This protocol ensures consistent, maintainable, and reliable tests across the project.