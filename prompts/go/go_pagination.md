<anthropic_go_pagination_protocol>

## Core Principles

1. OFFSET Pagination is Prohibited
- Do not use LIMIT + OFFSET pagination
- Special cases must be documented with clear parameter limits
- Must include explicit documentation for any exceptions

2. Cursor-Based Pagination is Required
- More consistent with live data updates
- Better performance on large datasets
- Prevents skipping or duplicating data

## Database Implementation

### Basic Cursor Pagination
```sql
-- Good: Cursor-based pagination
SELECT id, name 
FROM users
WHERE id > $lastId
ORDER BY id ASC
LIMIT $limit;

-- Bad: Offset-based pagination (prohibited)
SELECT id, name
FROM users
ORDER BY id ASC
OFFSET $offset
LIMIT $limit;
```

### Compound Cursors
For non-sequential IDs (e.g., UUIDs), use compound cursors:

```sql
-- Index for compound cursor
CREATE INDEX users_pagination ON users (
    created_at ASC,
    id ASC
);

-- Query with compound cursor
SELECT id, name, created_at
FROM users
WHERE 
    (created_at = $cursorTime AND id > $cursorId)
    OR created_at > $cursorTime
ORDER BY 
    created_at ASC,
    id ASC
LIMIT $limit;
```

### Filtered Pagination
When adding filters, include them in the index:

```sql
-- Index for filtered pagination
CREATE INDEX users_pagination ON users (
    created_at ASC,
    id ASC,
    status ASC
);

-- Query with filters
SELECT id, name, created_at
FROM users
WHERE 
    (
        (created_at = $cursorTime AND id > $cursorId)
        OR created_at > $cursorTime
    )
    AND status = 'active'
ORDER BY 
    created_at ASC,
    id ASC
LIMIT $limit;
```

## GraphQL Implementation

### Connection Model

#### Complete Mode (with Edges)
```graphql
type Query {
    users(first: Int!, after: String): UserConnection!
}

type UserConnection {
    totalCount: Int!
    edges: [UserEdge!]!
    pageInfo: PageInfo!
}

type UserEdge {
    node: User!
    cursor: String!
}

type PageInfo {
    startCursor: String!
    endCursor: String!
    hasNextPage: Boolean!
    hasPreviousPage: Boolean!
}
```

#### Simple Mode (without Edges)
```graphql
type Query {
    users(first: Int!, after: String): UserConnection!
}

type UserConnection {
    totalCount: Int!
    list: [User!]!
    pageInfo: PageInfo!
}
```

### Implementation Example

```go
type PageInfo struct {
    StartCursor     string `json:"startCursor"`
    EndCursor       string `json:"endCursor"`
    HasNextPage     bool   `json:"hasNextPage"`
    HasPreviousPage bool   `json:"hasPreviousPage"`
}

type UserConnection struct {
    TotalCount int       `json:"totalCount"`
    List       []*User   `json:"list"`
    PageInfo   PageInfo  `json:"pageInfo"`
}

func (r *queryResolver) Users(ctx context.Context, first int, after *string) (*UserConnection, error) {
    // Validate pagination parameters
    if first <= 0 || first > 100 {
        return nil, fmt.Errorf("first must be between 1 and 100")
    }

    // Decode cursor
    var cursor *time.Time
    if after != nil {
        decoded, err := DecodeCursor(*after)
        if err != nil {
            return nil, err
        }
        cursor = &decoded
    }

    // Query with one extra item to determine hasNextPage
    limit := first + 1
    users, err := r.db.QueryUsers(ctx, cursor, limit)
    if err != nil {
        return nil, err
    }

    // Determine if there are more pages
    hasNextPage := len(users) > first
    if hasNextPage {
        users = users[:first]
    }

    // Build connection
    edges := make([]*User, len(users))
    copy(edges, users)

    var startCursor, endCursor string
    if len(edges) > 0 {
        startCursor = EncodeCursor(edges[0].CreatedAt)
        endCursor = EncodeCursor(edges[len(edges)-1].CreatedAt)
    }

    return &UserConnection{
        TotalCount: len(edges),
        List:      edges,
        PageInfo: PageInfo{
            StartCursor:     startCursor,
            EndCursor:       endCursor,
            HasNextPage:     hasNextPage,
            HasPreviousPage: cursor != nil,
        },
    }, nil
}
```

## Best Practices

### Performance
1. Always create appropriate indexes
2. Use compound cursors for non-sequential IDs
3. Include filters in indexes
4. Limit maximum page size

### Error Handling
1. Validate pagination parameters
2. Handle cursor encoding/decoding errors
3. Provide clear error messages
4. Consider rate limiting for large datasets

### Edge Cases
1. Handle empty results
2. Deal with deleted items
3. Consider concurrent updates
4. Handle invalid cursors

### Implementation Guidelines
1. Always fetch limit+1 to determine hasNextPage
2. Use base64 for cursor encoding
3. Keep cursors opaque to clients
4. Consider caching for frequently accessed pages

> Claude must follow this protocol in implementing pagination.

</anthropic_go_pagination_protocol>