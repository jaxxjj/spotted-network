<anthropic_go_unit_test_protocol>

## Core Principles

1. Test Structure
- Use descriptive test names
- Follow AAA pattern (Arrange, Act, Assert)
- Keep tests independent and isolated
- Write maintainable test code

2. Test Coverage
- Aim for high test coverage but focus on critical paths
- Test both success and failure cases
- Test edge cases and boundary conditions
- Consider performance implications

## Testing Frameworks

### Using testify/suite
```go
type MySuite struct {
    suite.Suite
    db  *sql.DB
    svc *service
}

func TestMySuite(t *testing.T) {
    suite.Run(t, new(MySuite))
}

func (s *MySuite) SetupSuite() {
    // Suite-wide setup
}

func (s *MySuite) TearDownSuite() {
    // Suite-wide teardown
}

func (s *MySuite) SetupTest() {
    // Per-test setup
}

func (s *MySuite) TearDownTest() {
    // Per-test cleanup
}
```

### Using wpgx/testsuite for Database Testing
```go
type MyDBSuite struct {
    *testsuite.WPgxTestSuite
    svc *service
}

func newMyDBSuite() *MyDBSuite {
    return &MyDBSuite{
        WPgxTestSuite: testsuite.NewWPgxTestSuiteFromEnv(
            "test_db",
            []string{
                // Schema definitions
                `CREATE TABLE users (
                    id SERIAL PRIMARY KEY,
                    name TEXT NOT NULL
                );`,
                // Add more schema definitions
            },
        ),
    }
}

func TestMyDBSuite(t *testing.T) {
    suite.Run(t, newMyDBSuite())
}

func (s *MyDBSuite) SetupTest() {
    s.WPgxTestSuite.SetupTest()
    // Additional setup
    s.svc = NewService(s.GetPool())
}
```

## Test Patterns

### Table-Driven Tests
```go
func TestCalculator(t *testing.T) {
    tests := map[string]struct {
        x, y   int
        op     string
        want   int
        errMsg string
    }{
        "simple addition": {
            x:    2,
            y:    2,
            op:   "+",
            want: 4,
        },
        "division by zero": {
            x:      1,
            y:      0,
            op:     "/",
            errMsg: "division by zero",
        },
    }

    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            got, err := Calculate(tc.x, tc.y, tc.op)
            if tc.errMsg != "" {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tc.errMsg)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tc.want, got)
        })
    }
}
```

### Database Tests with Transactions
```go
func (s *MyDBSuite) TestUserCreation() {
    tests := map[string]struct {
        input    User
        wantErr  bool
        errMsg   string
    }{
        "valid user": {
            input: User{Name: "Alice"},
        },
        "empty name": {
            input:   User{Name: ""},
            wantErr: true,
            errMsg:  "name cannot be empty",
        },
    }

    for name, tc := range tests {
        s.Run(name, func() {
            // Each test runs in a transaction
            err := s.svc.CreateUser(s.Ctx(), tc.input)
            if tc.wantErr {
                s.Error(err)
                s.Contains(err.Error(), tc.errMsg)
                return
            }
            s.NoError(err)
            
            // Verify user was created
            var user User
            err = s.GetPool().QueryRow(s.Ctx(), 
                "SELECT name FROM users WHERE id = $1", 
                tc.input.ID,
            ).Scan(&user.Name)
            s.NoError(err)
            s.Equal(tc.input.Name, user.Name)
        })
    }
}
```

### Concurrent Tests
```go
func TestConcurrent(t *testing.T) {
    t.Parallel() // Mark test as parallel-capable
    
    tests := map[string]struct {
        input    string
        want     string
    }{
        "test1": {input: "a", want: "A"},
        "test2": {input: "b", want: "B"},
    }
    
    for name, tc := range tests {
        name, tc := name, tc // Capture range variables
        t.Run(name, func(t *testing.T) {
            t.Parallel() // Run subtests in parallel
            got := process(tc.input)
            assert.Equal(t, tc.want, got)
        })
    }
}
```

### Mocking Dependencies
```go
type mockDB struct {
    mock.Mock
}

func (m *mockDB) GetUser(ctx context.Context, id int) (*User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func TestService(t *testing.T) {
    mockDB := new(mockDB)
    svc := NewService(mockDB)
    
    // Setup expectations
    mockDB.On("GetUser", mock.Anything, 1).Return(&User{
        ID: 1,
        Name: "Test",
    }, nil)
    
    // Test
    user, err := svc.GetUser(context.Background(), 1)
    require.NoError(t, err)
    assert.Equal(t, "Test", user.Name)
    
    // Verify expectations
    mockDB.AssertExpectations(t)
}
```

## Best Practices

1. Test Organization
- Group related tests in test suites
- Use meaningful test names
- Keep test files next to source files
- Follow consistent naming conventions

2. Test Data
- Use table-driven tests for multiple test cases
- Create test helpers for common setup
- Use meaningful test data
- Clean up test data in teardown

3. Assertions
- Use testify/require for critical assertions
- Use testify/assert for non-critical checks
- Write clear failure messages
- Check error types and messages

4. Database Testing
- Use transactions for isolation
- Clean up test data
- Test database constraints
- Consider performance impact

5. Mocking
- Mock at interface boundaries
- Only mock what's necessary
- Verify mock expectations
- Keep mocks simple

6. Error Handling
- Test both success and error paths
- Verify error types and messages
- Test timeout scenarios
- Test cancellation scenarios

7. Performance
- Use t.Parallel() when appropriate
- Benchmark critical paths
- Profile test execution
- Monitor resource usage

## Common Patterns

### Repository Tests
```go
func (s *MyDBSuite) TestUserRepo() {
    // Create test data
    user := &User{Name: "Test"}
    err := s.repo.Create(s.Ctx(), user)
    s.NoError(err)
    
    // Read and verify
    found, err := s.repo.GetByID(s.Ctx(), user.ID)
    s.NoError(err)
    s.Equal(user.Name, found.Name)
    
    // Update
    user.Name = "Updated"
    err = s.repo.Update(s.Ctx(), user)
    s.NoError(err)
    
    // Delete
    err = s.repo.Delete(s.Ctx(), user.ID)
    s.NoError(err)
    
    // Verify deletion
    _, err = s.repo.GetByID(s.Ctx(), user.ID)
    s.Error(err) // Should return error
}
```

### Service Tests
```go
func (s *MySuite) TestUserService() {
    // Setup mocks
    mockRepo := new(MockUserRepo)
    mockRepo.On("GetByID", mock.Anything, 1).Return(&User{
        ID: 1,
        Name: "Test",
    }, nil)
    
    svc := NewUserService(mockRepo)
    
    // Test service method
    user, err := svc.GetUser(context.Background(), 1)
    s.NoError(err)
    s.Equal("Test", user.Name)
    
    // Verify mock
    mockRepo.AssertExpectations(s.T())
}
```

### API Tests
```go
func (s *MySuite) TestUserAPI() {
    // Setup test server
    handler := NewUserHandler(s.svc)
    server := httptest.NewServer(handler)
    defer server.Close()
    
    // Make request
    resp, err := http.Get(server.URL + "/users/1")
    s.NoError(err)
    defer resp.Body.Close()
    
    // Verify response
    s.Equal(http.StatusOK, resp.StatusCode)
    var user User
    err = json.NewDecoder(resp.Body).Decode(&user)
    s.NoError(err)
    s.Equal("Test", user.Name)
}
```

> Claude must follow these patterns when implementing Go unit tests.

</anthropic_go_unit_test_protocol>