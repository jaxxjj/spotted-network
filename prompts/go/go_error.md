<anthropic_go_error_protocol>

## Core Concepts

### Error Model
Galxe errors follow the gRPC status error structure with three key components:
1. Status Code - Standard gRPC status codes
2. Domain - Source of the error
3. Details - Error specifics including reason and arguments

### Error Priority Flow
Frontend error display priority:
1. Domain + Reason -> Load from local language JSON
2. Domain + Reason -> Fallback to English JSON
3. Raw message if no i18n definition exists

## Basic Usage

### NewError Function
```go
func NewError(
    code codes.Code,      // gRPC status code
    domain Domain,        // Error domain
    originalError error,  // Original error
    reason string,        // Error reason key
    args ...interface{},  // Message arguments
) error
```

### Example Usage
```go
// Basic error with reason
if err != nil {
    return galerr.NewError(
        codes.NotFound,           // gRPC code
        galerr.DOMAIN_TWITTER,    // Domain
        err,                      // Original error
        "ENT_NOT_FOUND",         // Reason key
        "credential", credID,     // Arguments
    )
}

// Error response structure
{
    "error": {
        "message": "cannot find ent entity: credential, id: 123.[ERROR: ent: credential not found]",
        "status": "INVALID_ARGUMENT",
        "details": [{
            "domain": "TWITTER",
            "reason": "ENT_NOT_FOUND",
            "args": ["credential", 123]
        }]
    }
}
```

## Best Practices

### 1. Error Source Handling

#### Internal Service Errors
```go
// Return gRPC service errors directly
if err != nil {
    return err
}
```

#### External Service Errors
```go
// Wrap third-party errors with appropriate domain
if err != nil {
    return galerr.NewError(
        codes.Internal,
        galerr.DOMAIN_TWITTER,
        err,                // Original error preserved
        "API_ERROR"        // Standard reason key
    )
}
```

#### Rate Limiting Errors
```go
// Clean error message for rate limits
if isRateLimited {
    return galerr.NewError(
        codes.ResourceExhausted,
        galerr.DOMAIN_TWITTER,
        nil,
        "PLEASE_WAIT_IN_MIN",
        10  // Minutes to wait
    )
}

// With original error included
if err != nil {
    return galerr.NewError(
        codes.ResourceExhausted,
        galerr.DOMAIN_TWITTER,
        err,  // Will append [ERROR: original message]
        "PLEASE_WAIT_IN_MIN",
        10
    )
}
```

### 2. Status Code Selection

Use appropriate gRPC status codes:
- `codes.InvalidArgument` - Bad input parameters
- `codes.NotFound` - Resource not found
- `codes.Internal` - Internal server errors
- `codes.ResourceExhausted` - Rate limiting/quotas
- `codes.Unauthenticated` - Authentication required
- `codes.PermissionDenied` - Authorization failed

### 3. Error Message Design

#### Use Constants for Reason Keys
```go
// Bad: String literals
return galerr.NewError(codes.NotFound, domain, err, "user_not_found")

// Good: Generated constants
return galerr.NewError(codes.NotFound, domain, err, galerr.BackendUserNotFound)
```

#### Message Templates
- Use placeholder numbering: "{0}, {1}"
- Keep messages concise and clear
- Include actionable information
- Consider i18n requirements

### 4. Domain Organization

#### Register New Domains
```go
// In domain.go
const (
    DOMAIN_BACKEND  Domain = "backend"
    DOMAIN_TWITTER  Domain = "twitter"
    // Add new domains here
)
```

#### Organize Reason Files
```
pkg/
  reasons/
    twitter/
      en.json
      zh.json
    backend/
      en.json
      zh.json
```

### 5. Error Propagation

#### Service Layer
```go
func (s *service) DoSomething(ctx context.Context) error {
    // Propagate repository errors with domain
    user, err := s.repo.GetUser(ctx, id)
    if err != nil {
        return galerr.NewError(
            codes.Internal,
            galerr.DOMAIN_BACKEND,
            err,
            "USER_FETCH_ERROR"
        )
    }
    
    // Business logic errors
    if !user.IsActive {
        return galerr.NewError(
            codes.FailedPrecondition,
            galerr.DOMAIN_BACKEND,
            nil,
            "USER_INACTIVE"
        )
    }
    
    return nil
}
```

#### API Layer
```go
func (s *Server) HandleRequest(ctx context.Context, req *pb.Request) (*pb.Response, error) {
    // Validate input
    if req.ID == "" {
        return nil, galerr.NewError(
            codes.InvalidArgument,
            galerr.DOMAIN_BACKEND,
            nil,
            "INVALID_ID"
        )
    }
    
    // Propagate service errors
    if err := s.svc.DoSomething(ctx); err != nil {
        return nil, err  // Already wrapped, just propagate
    }
    
    return &pb.Response{}, nil
}
```

## Setup Guide

1. Add Domain (if needed)
```go
// pkg/domain.go
const DOMAIN_NEWSERVICE Domain = "newservice"
```

2. Create Reason Files
```json
// pkg/reasons/newservice/en.json
{
    "RESOURCE_NOT_FOUND": "Cannot find {0} with ID {1}",
    "RATE_LIMITED": "Please wait for {0} minutes"
}
```

3. Generate Error Keys
```shell
make errkeygen
```

4. Import and Use
```go
import galerr "github.com/NFTGalaxy/galaxy-errors/pkg"

func main() {
    err := galerr.NewError(
        codes.NotFound,
        galerr.DOMAIN_NEWSERVICE,
        nil,
        galerr.NewserviceResourceNotFound,
        "user",
        123
    )
}
```

## Common Patterns

### 1. Database Errors
```go
// Repository layer
func (r *repo) Get(ctx context.Context, id int) (*Entity, error) {
    entity, err := r.db.Get(ctx, id)
    if ent.IsNotFound(err) {
        return nil, galerr.NewError(
            codes.NotFound,
            galerr.DOMAIN_BACKEND,
            err,
            "ENTITY_NOT_FOUND",
            id
        )
    }
    if err != nil {
        return nil, galerr.NewError(
            codes.Internal,
            galerr.DOMAIN_BACKEND,
            err,
            "DATABASE_ERROR"
        )
    }
    return entity, nil
}
```

### 2. External API Errors
```go
// Client layer
func (c *client) FetchData(ctx context.Context) error {
    resp, err := c.httpClient.Get(url)
    if err != nil {
        return galerr.NewError(
            codes.Internal,
            galerr.DOMAIN_TWITTER,
            err,
            "API_REQUEST_FAILED"
        )
    }
    
    if resp.StatusCode == http.StatusTooManyRequests {
        return galerr.NewError(
            codes.ResourceExhausted,
            galerr.DOMAIN_TWITTER,
            nil,
            "RATE_LIMITED",
            c.getWaitTime()
        )
    }
    
    return nil
}
```

### 3. Validation Errors
```go
func validateUser(user *User) error {
    if user.Name == "" {
        return galerr.NewError(
            codes.InvalidArgument,
            galerr.DOMAIN_BACKEND,
            nil,
            "REQUIRED_FIELD",
            "name"
        )
    }
    
    if len(user.Password) < 8 {
        return galerr.NewError(
            codes.InvalidArgument,
            galerr.DOMAIN_BACKEND,
            nil,
            "INVALID_PASSWORD_LENGTH",
            8
        )
    }
    
    return nil
}
```

> Claude must follow these patterns when implementing error handling in Go services.

## GraphQL Error Handling

### Query Resolver
Query resolver should return errors directly to fail fast:
```go
func (r *queryResolver) Campaign(ctx context.Context, id string) (*model.Campaign, error) {
    campId, err := util.CampaignID2Int(id)
    if err != nil {
        return nil, galerr.NewError(
            codes.InvalidArgument,
            galerr.DOMAIN_BACKEND,
            err,
            "INVALID_CAMPAIGN_ID",
            id
        )
    }
    // ...
}
```

### Field Resolver
Field resolvers should use AppendError to allow partial data return:
```go
func (r *campaignResolver) IsBookmarked(ctx context.Context, obj *model.Campaign, address string) (bool, error) {
    bookmarkSet, err := r.getBookmarks(ctx, address)
    if err != nil {
        // Use AppendError instead of returning error directly
        AppendError(ctx, galerr.NewError(
            codes.Internal,
            galerr.DOMAIN_BACKEND,
            err,
            "BOOKMARK_FETCH_ERROR"
        ))
        return false, nil
    }
    return bookmarkSet.isBookmarked, nil
}

// Bad Practice - Don't return error directly in field resolver
func (r *campaignResolver) NumNFTMinted(ctx context.Context, obj *model.Campaign) (*int, error) {
    id, _ := util.CampaignID2Int(obj.ID)
    resp, err := r.client.GetCampaignNumNftMinted(ctx, &backend.GetCampaignNumNftMintedRequest{
        CampaignId: id,
    })
    if err != nil {
        return nil, err // Should use AppendError instead
    }
    conv := int(resp.GetNumNftMinted())
    return &conv, nil
}
```

### Mutation Resolver
Mutations should return errors directly:
```go
func (r *mutationResolver) UpdateCampaign(ctx context.Context, input model.UpdateCampaignInput) (*model.Campaign, error) {
    if input.Name == "" {
        return nil, galerr.NewError(
            codes.InvalidArgument,
            galerr.DOMAIN_BACKEND,
            nil,
            "REQUIRED_FIELD",
            "name"
        )
    }
    
    campaign, err := r.svc.UpdateCampaign(ctx, input)
    if err != nil {
        return nil, err // Already wrapped by service layer
    }
    
    return campaign, nil
}
```

## Engineering Practices

### Error Package Management
Use git submodules to manage error package:
```
// .gitmodules
[submodule "proto/errors"]
    path = proto/errors
    url = git@github.com:NFTGalaxy/galaxy-errors.git
```

### Makefile Error Handling
Include error checking in Makefile targets:
```makefile
.PHONY: build test lint codecov

build:
    go build -v ./...
    
test:
    go test -race -coverprofile=coverage.txt -covermode=atomic ./...
    
lint:
    golangci-lint run
    
codecov:
    bash <(curl -s https://codecov.io/bash)
```

### Graceful Shutdown
Use error groups for coordinated shutdown:
```go
func (a *App) Start() error {
    errGroup, ctx := errgroup.WithContext(context.Background())
    
    // Start services
    errGroup.Go(func() error {
        if err := a.grpcServer.Start(ctx); err != nil {
            return galerr.NewError(
                codes.Internal,
                galerr.DOMAIN_BACKEND,
                err,
                "GRPC_SERVER_ERROR"
            )
        }
        return nil
    })
    
    errGroup.Go(func() error {
        if err := a.httpServer.Start(ctx); err != nil {
            return galerr.NewError(
                codes.Internal,
                galerr.DOMAIN_BACKEND,
                err,
                "HTTP_SERVER_ERROR"
            )
        }
        return nil
    })
    
    // Handle shutdown signals
    errGroup.Go(func() error {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        select {
        case <-ctx.Done():
            return ctx.Err()
        case sig := <-sigCh:
            return fmt.Errorf("received signal: %v", sig)
        }
    })
    
    return errGroup.Wait()
}
```

### Testing
1. Local gRPC Testing
```go
func TestCampaignService(t *testing.T) {
    // Setup test server
    srv := grpc.NewServer()
    pb.RegisterCampaignServiceServer(srv, NewCampaignService())
    
    // Test cases
    tests := map[string]struct {
        req     *pb.GetCampaignRequest
        wantErr *galerr.Error
    }{
        "invalid id": {
            req: &pb.GetCampaignRequest{Id: "invalid"},
            wantErr: galerr.NewError(
                codes.InvalidArgument,
                galerr.DOMAIN_BACKEND,
                nil,
                "INVALID_CAMPAIGN_ID",
            ),
        },
    }
    
    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            _, err := srv.GetCampaign(context.Background(), tc.req)
            if tc.wantErr != nil {
                require.Error(t, err)
                st, ok := status.FromError(err)
                require.True(t, ok)
                require.Equal(t, tc.wantErr.Code(), st.Code())
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

2. Error Response Testing
```go
func TestGraphQLErrors(t *testing.T) {
    srv := handler.NewDefaultServer(
        generated.NewExecutableSchema(
            generated.Config{
                Resolvers: &Resolver{},
            },
        ),
    )
    
    tests := map[string]struct {
        query   string
        wantErr string
    }{
        "invalid campaign": {
            query: `
                query {
                    campaign(id: "invalid") {
                        id
                        name
                    }
                }
            `,
            wantErr: "INVALID_CAMPAIGN_ID",
        },
    }
    
    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            resp := srv.Schema.Exec(context.Background(), tc.query)
            require.NotNil(t, resp.Errors)
            require.Contains(t, resp.Errors[0].Message, tc.wantErr)
        })
    }
}
```

> Claude must follow this protocol in implementing Go errors.

</anthropic_graphql_error_protocol>