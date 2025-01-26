<anthropic_sqlc_protocol>

## Core Architecture and Setup

```yaml
version: '2'
sql:
  - schema: books/schema.sql    # Primary schema file
    queries: books/query.sql    # Query definitions
    engine: postgresql
    gen:
      go:
        sql_package: wpgx      # Must use wpgx
        package: books
        out: books
```

## Schema Design Guidelines

### Schema File Organization

One schema file per logical table/view
Include all constraints and indexes
Always use IF NOT EXISTS clauses
Follow established naming conventions (snake_case)

Example:
```sql
CREATE TABLE IF NOT EXISTS books (
   id          BIGSERIAL       GENERATED ALWAYS AS IDENTITY,
   name        VARCHAR(255)    NOT NULL,
   created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
   CONSTRAINT books_id_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS books_name_idx ON books (name);
```

### Schema Dependencies

List dependencies in topological order
Reference schemas must be listed after dependent schemas

```yaml
- schema:
    - orders/schema.sql    # Parent schema first
    - books/schema.sql     # Referenced schemas after
```

## Query Design

### Required Options
Every query must specify:
```sql
-- name: GetBook :one
-- -- timeout : 500ms      # Mandatory timeout
-- -- cache : 10m         # Optional cache duration
SELECT * FROM books WHERE id = @id;
```

### Parameter Styles
```sql
-- Best Practices:
- Use @param_name for named parameters
- One parameter style per query (@, $, or sqlc.arg)
- Use ::type for explicit type casting
```

### Bulk Operations
```sql
-- Bulk insert with UNNEST
-- name: BulkInsert :exec
INSERT INTO books (name, price)
SELECT 
  UNNEST(@names::VARCHAR[]),
  UNNEST(@prices::DECIMAL[]);
```

## Cache Strategy

### Caching Guidelines
```go
// Cache durations:
- Long cache (hours): Single-row queries with clear invalidation
- Short cache (minutes): Aggregate queries
- No cache: Real-time data requirements
```

### Invalidation Patterns
```sql
-- Execute cache invalidation after mutations
-- -- invalidate: GetBook
UPDATE books SET name = @name WHERE id = @id;
```

## Testing Framework

### Basic Test Structure
```go
type testSuite struct {
    *testsuite.WPgxTestSuite
}

func (s *testSuite) TestBookOperations() {
    // 1. Setup initial state
    s.LoadState("initial_books.json", s.booksSerde)
    
    // 2. Execute operations
    result, err := s.books.CreateBook(ctx, params)
    
    // 3. Verify state
    s.Golden("books_table", s.booksSerde)
}
```

### Golden Testing
```go
// Loading test data:
suite.LoadState("test_data.json", serde)

// Comparing states:
suite.Golden("table_name", serde)

// Template-based testing:
suite.LoadStateTmpl("template.json.tmpl", serde, timePoints)
```

## Best Practices

### Schema Practices
- Always use timestamptz for timestamps
- Include appropriate indexes
- Document constraints clearly
- Follow consistent naming conventions

### Query Practices
- Set appropriate timeouts
- Use explicit type casting
- Leverage bulk operations
- Consider cache implications

### Testing Practices
- Use golden files for state verification
- Template time-sensitive tests
- Reset sequence numbers after loading test data
- Normalize timestamps for comparisons

## Case Studies

### Bulk Operations

#### Simple Bulk Insert
- Use `:copyfrom` for simple bulk inserts without constraints
- Automatically rolls back on constraint violations
- Best performance for bulk inserts
```sql
-- name: BulkInsert :copyfrom
INSERT INTO books (
   name, description, metadata, category, price
) VALUES (
  $1, $2, $3, $4, $5
);
```

#### Bulk Upsert
- Uses `UNNEST` to handle array parameters
- Handles conflicts with `ON CONFLICT` clause
- Updates existing records if they exist
```sql
-- name: UpsertUsers :exec
insert into users (name, metadata, image)
select
  unnest(@name::VARCHAR(255)[]),
  unnest(@metadata::JSON[]),
  unnest(@image::TEXT[])
on conflict ON CONSTRAINT users_lower_name_key do
update set
    metadata = excluded.metadata,
    image = excluded.image;
```

#### Bulk Conditional Queries
- Uses `UNNEST` for efficient multiple conditions
- Better performance than multiple OR conditions
```sql
-- name: ListOrdersByUserAndBook :many
SELECT * FROM orders
WHERE (user_id, book_id) IN (
  SELECT
    UNNEST(@user_id::int[]),
    UNNEST(@book_id::int[])
);
```

#### Bulk Updates with Different Values
- Uses temporary table for matching updates
- More efficient than individual updates
```sql
-- name: BulkUpdate :exec
UPDATE orders
SET
  price=temp.price,
  book_id=temp.book_id
FROM (
  SELECT
    UNNEST(@id::int[]) as id,
    UNNEST(@price::bigint[]) as price,
    UNNEST(@book_id::int[]) as book_id
) AS temp
WHERE orders.id=temp.id;
```

### Advanced Patterns

#### Partial Updates
- Uses `sqlc.narg` for nullable parameters
- Uses `coalesce` for optional updates
- Cannot set nullable columns to null
```sql
-- name: PartialUpdateByID :exec
UPDATE books
SET
  description = coalesce(sqlc.narg('description'), description),
  metadata = coalesce(sqlc.narg('meta'), metadata),
  price = coalesce(sqlc.narg('price'), price),
  updated_at = NOW()
WHERE id = sqlc.arg('id');
```

#### Versatile Queries
- Combines multiple optional conditions
- Special handling for nullable columns
```sql
-- name: GetBookBySpec :one
-- -- cache : 10m
SELECT * FROM books WHERE
  name LIKE coalesce(sqlc.narg('name'), name) AND
  price = coalesce(sqlc.narg('price'), price) AND
  (sqlc.narg('dummy')::int is NULL or dummy_field = sqlc.narg('dummy'));
```

#### Materialized View Refresh
- `CONCURRENTLY` for non-blocking refreshes
- Requires unique index for concurrent refresh
- First refresh must be non-concurrent
```sql
-- name: Refresh :exec
REFRESH MATERIALIZED VIEW CONCURRENTLY by_book_revenues;
```

## Index Strategy

### Index Types
- B-Tree (Default): Suitable for most scenarios
  ```sql
  CREATE INDEX books_name_idx ON books(name);
  ```
- Hash: Only for equality comparisons, supported in PostgreSQL 10+
  ```sql
  CREATE INDEX books_isbn_idx ON books USING HASH (isbn);
  ```
- GIN: For array values and full-text search
  ```sql
  CREATE INDEX books_tags_idx ON books USING GIN (tags);
  ```
- GiST: For geometric data and special data types
  ```sql
  CREATE INDEX geo_idx ON locations USING GIST (position);
  ```

### Special Index Types

#### Partial Index
- Index only rows matching a condition, reduces index size
```sql
-- Only index published books
CREATE INDEX books_published_idx ON books(published_at) 
WHERE status = 'published';
```

#### Expression Index
- Index results of expressions instead of simple columns
- Useful for computed value searches
- More expensive to maintain but faster for retrieval
```sql
-- Case-insensitive search
CREATE INDEX books_lower_name_idx ON books(LOWER(name));

-- Date search optimization
CREATE INDEX books_date_idx ON books(DATE(created_at));

-- Concatenated fields search
CREATE INDEX books_full_title_idx ON books((title || ' ' || subtitle));

-- Complex expressions (use parentheses)
CREATE INDEX books_complex_idx ON books((price * discount));
```

Key considerations:
- Maintenance cost: Computed for each insert/update
- Search benefit: Expression not recomputed during queries
- Use when retrieval speed > update speed
- Can enforce constraints not possible with simple unique indexes

#### Unique Index
- Ensures data uniqueness and improves query performance
```sql
-- Ensure ISBN uniqueness
CREATE UNIQUE INDEX books_isbn_unique ON books(isbn);
-- Partial unique index
CREATE UNIQUE INDEX books_name_category_idx ON books(name, category) 
WHERE status = 'active';
```

#### Multi-Column Index
- Optimizes queries with multiple column conditions
```sql
-- Optimizes WHERE author_id = x AND category = y
CREATE INDEX books_author_category_idx ON books(author_id, category);
```

### Index Maintenance

#### Concurrent Creation
- Create indexes without blocking writes
```sql
-- Create index concurrently
CREATE INDEX CONCURRENTLY books_title_idx ON books(title);
```

#### Index Optimization
- Rebuild indexes to optimize performance
```sql
-- Rebuild single index
REINDEX INDEX books_title_idx;
-- Rebuild all indexes on table
REINDEX TABLE books;
```

### Best Practices
1. Selectivity Principles
   - Index columns with high selectivity
   - Avoid indexing small tables

2. Maintenance Guidelines
   - Use CONCURRENTLY for production index creation
   - Regularly rebuild indexes on frequently updated tables
   - Remove unused indexes

3. Performance Considerations
   - Prefer B-tree indexes for general use
   - Consider partial indexes to reduce size
   - Use multi-column indexes judiciously

> Claude must follow this protocol in implementing sqlc and go files.

</anthropic_sqlc_protocol>