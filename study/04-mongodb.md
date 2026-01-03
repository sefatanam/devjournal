# MongoDB Database

## Overview

DevJournal uses MongoDB for document storage where flexibility is needed. Code snippets are stored in MongoDB because:

- **Flexible metadata** - Different snippets may have different properties
- **Nested structures** - Code examples with multiple files
- **Schema evolution** - Easy to add new fields without migrations
- **Text search** - Built-in full-text indexing

## What's Stored in MongoDB?

| Collection | Purpose | Why MongoDB? |
|------------|---------|--------------|
| `snippets` | Code snippets | Flexible metadata, varying structures |

## Document Schema

### Snippet Document

```javascript
{
  "_id": ObjectId("65a1b2c3d4e5f6789012345"),
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Angular Signal Store Example",
  "description": "Complete example of NgRx Signal Store",
  "code": "import { signalStore } from '@ngrx/signals';\n...",
  "language": "typescript",
  "tags": ["angular", "signals", "state-management"],
  "metadata": {
    "framework": "angular",
    "version": "19",
    "category": "state-management"
  },
  "is_favorite": true,
  "view_count": 42,
  "created_at": ISODate("2024-01-15T10:30:00Z"),
  "updated_at": ISODate("2024-01-15T10:30:00Z")
}
```

## Go Domain Model

```go
// services/go-api/internal/domain/snippet.go

package domain

import "time"

type Snippet struct {
    ID          string            `bson:"_id,omitempty" json:"id"`
    UserID      string            `bson:"user_id" json:"userId"`
    Title       string            `bson:"title" json:"title"`
    Description string            `bson:"description" json:"description"`
    Code        string            `bson:"code" json:"code"`
    Language    string            `bson:"language" json:"language"`
    Tags        []string          `bson:"tags" json:"tags"`
    Metadata    map[string]string `bson:"metadata,omitempty" json:"metadata,omitempty"`
    IsFavorite  bool              `bson:"is_favorite" json:"isFavorite"`
    ViewCount   int               `bson:"view_count" json:"viewCount"`
    CreatedAt   time.Time         `bson:"created_at" json:"createdAt"`
    UpdatedAt   time.Time         `bson:"updated_at" json:"updatedAt"`
}

type SnippetFilter struct {
    Page     int
    PageSize int
    Language string
    Tag      string
    Search   string
    Favorite bool
}
```

## MongoDB Connection

```go
// services/go-api/internal/database/mongodb.go

package database

import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDB struct {
    Client   *mongo.Client
    Database *mongo.Database
}

func NewMongoDB(ctx context.Context, uri, dbName string) (*MongoDB, error) {
    // Configure client options
    clientOptions := options.Client().
        ApplyURI(uri).
        SetMinPoolSize(10).
        SetMaxPoolSize(50).
        SetMaxConnIdleTime(5 * time.Minute).
        SetServerSelectionTimeout(5 * time.Second)

    // Connect to MongoDB
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
    }

    // Ping to verify connection
    if err := client.Ping(ctx, readpref.Primary()); err != nil {
        return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
    }

    return &MongoDB{
        Client:   client,
        Database: client.Database(dbName),
    }, nil
}

func (m *MongoDB) Close(ctx context.Context) error {
    return m.Client.Disconnect(ctx)
}

// Collection returns a specific collection
func (m *MongoDB) Collection(name string) *mongo.Collection {
    return m.Database.Collection(name)
}
```

## Repository Implementation

```go
// services/go-api/internal/repository/mongodb/snippet_repo.go

package mongodb

import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type SnippetRepo struct {
    collection *mongo.Collection
}

func NewSnippetRepo(db *database.MongoDB) *SnippetRepo {
    repo := &SnippetRepo{
        collection: db.Collection("snippets"),
    }
    // Create indexes on initialization
    repo.createIndexes(context.Background())
    return repo
}

// createIndexes sets up MongoDB indexes
func (r *SnippetRepo) createIndexes(ctx context.Context) error {
    indexes := []mongo.IndexModel{
        // Compound index for user queries
        {
            Keys: bson.D{
                {Key: "user_id", Value: 1},
                {Key: "created_at", Value: -1},
            },
        },
        // Index for language filtering
        {
            Keys: bson.D{{Key: "language", Value: 1}},
        },
        // Index for tags
        {
            Keys: bson.D{{Key: "tags", Value: 1}},
        },
        // Text index for search
        {
            Keys: bson.D{
                {Key: "title", Value: "text"},
                {Key: "description", Value: "text"},
                {Key: "code", Value: "text"},
            },
            Options: options.Index().SetWeights(bson.D{
                {Key: "title", Value: 10},       // Title has highest weight
                {Key: "description", Value: 5},  // Description medium weight
                {Key: "code", Value: 1},         // Code lowest weight
            }),
        },
    }

    _, err := r.collection.Indexes().CreateMany(ctx, indexes)
    return err
}

// Create inserts a new snippet
func (r *SnippetRepo) Create(ctx context.Context, snippet *domain.Snippet) (*domain.Snippet, error) {
    snippet.ID = primitive.NewObjectID().Hex()
    snippet.CreatedAt = time.Now()
    snippet.UpdatedAt = time.Now()
    snippet.ViewCount = 0

    _, err := r.collection.InsertOne(ctx, snippet)
    if err != nil {
        return nil, fmt.Errorf("failed to create snippet: %w", err)
    }

    return snippet, nil
}

// GetByID retrieves a snippet by ID
func (r *SnippetRepo) GetByID(ctx context.Context, id string) (*domain.Snippet, error) {
    var snippet domain.Snippet

    err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&snippet)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("failed to get snippet: %w", err)
    }

    return &snippet, nil
}

// List retrieves paginated snippets with filters
func (r *SnippetRepo) List(
    ctx context.Context,
    userID string,
    filter domain.SnippetFilter,
) ([]*domain.Snippet, int, error) {
    // Build filter query
    query := bson.M{"user_id": userID}

    if filter.Language != "" {
        query["language"] = filter.Language
    }

    if filter.Tag != "" {
        query["tags"] = filter.Tag // MongoDB handles array contains automatically
    }

    if filter.Favorite {
        query["is_favorite"] = true
    }

    if filter.Search != "" {
        query["$text"] = bson.M{"$search": filter.Search}
    }

    // Count total documents
    total, err := r.collection.CountDocuments(ctx, query)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to count snippets: %w", err)
    }

    // Configure find options
    skip := int64((filter.Page - 1) * filter.PageSize)
    limit := int64(filter.PageSize)

    findOptions := options.Find().
        SetSkip(skip).
        SetLimit(limit).
        SetSort(bson.D{{Key: "created_at", Value: -1}})

    // If searching, sort by text score
    if filter.Search != "" {
        findOptions.SetProjection(bson.M{
            "score": bson.M{"$meta": "textScore"},
        })
        findOptions.SetSort(bson.D{{Key: "score", Value: bson.M{"$meta": "textScore"}}})
    }

    // Execute query
    cursor, err := r.collection.Find(ctx, query, findOptions)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to list snippets: %w", err)
    }
    defer cursor.Close(ctx)

    // Decode results
    var snippets []*domain.Snippet
    if err := cursor.All(ctx, &snippets); err != nil {
        return nil, 0, fmt.Errorf("failed to decode snippets: %w", err)
    }

    return snippets, int(total), nil
}

// Update modifies an existing snippet
func (r *SnippetRepo) Update(ctx context.Context, snippet *domain.Snippet) (*domain.Snippet, error) {
    snippet.UpdatedAt = time.Now()

    update := bson.M{
        "$set": bson.M{
            "title":       snippet.Title,
            "description": snippet.Description,
            "code":        snippet.Code,
            "language":    snippet.Language,
            "tags":        snippet.Tags,
            "metadata":    snippet.Metadata,
            "is_favorite": snippet.IsFavorite,
            "updated_at":  snippet.UpdatedAt,
        },
    }

    result, err := r.collection.UpdateOne(
        ctx,
        bson.M{"_id": snippet.ID, "user_id": snippet.UserID},
        update,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to update snippet: %w", err)
    }

    if result.MatchedCount == 0 {
        return nil, ErrNotFound
    }

    return snippet, nil
}

// Delete removes a snippet
func (r *SnippetRepo) Delete(ctx context.Context, id, userID string) error {
    result, err := r.collection.DeleteOne(ctx, bson.M{
        "_id":     id,
        "user_id": userID,
    })
    if err != nil {
        return fmt.Errorf("failed to delete snippet: %w", err)
    }

    if result.DeletedCount == 0 {
        return ErrNotFound
    }

    return nil
}

// IncrementViewCount atomically increments view count
func (r *SnippetRepo) IncrementViewCount(ctx context.Context, id string) error {
    _, err := r.collection.UpdateOne(
        ctx,
        bson.M{"_id": id},
        bson.M{"$inc": bson.M{"view_count": 1}},
    )
    return err
}

// GetLanguageStats returns snippet counts by language
func (r *SnippetRepo) GetLanguageStats(ctx context.Context, userID string) ([]domain.LanguageStat, error) {
    pipeline := mongo.Pipeline{
        // Match user's snippets
        {{Key: "$match", Value: bson.M{"user_id": userID}}},
        // Group by language and count
        {{Key: "$group", Value: bson.M{
            "_id":   "$language",
            "count": bson.M{"$sum": 1},
        }}},
        // Sort by count descending
        {{Key: "$sort", Value: bson.M{"count": -1}}},
    }

    cursor, err := r.collection.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, fmt.Errorf("failed to aggregate stats: %w", err)
    }
    defer cursor.Close(ctx)

    var results []struct {
        Language string `bson:"_id"`
        Count    int    `bson:"count"`
    }
    if err := cursor.All(ctx, &results); err != nil {
        return nil, fmt.Errorf("failed to decode stats: %w", err)
    }

    stats := make([]domain.LanguageStat, len(results))
    for i, r := range results {
        stats[i] = domain.LanguageStat{
            Language: r.Language,
            Count:    r.Count,
        }
    }

    return stats, nil
}

// ToggleFavorite toggles the favorite status
func (r *SnippetRepo) ToggleFavorite(ctx context.Context, id, userID string) (bool, error) {
    // First get current state
    var snippet domain.Snippet
    err := r.collection.FindOne(ctx, bson.M{"_id": id, "user_id": userID}).Decode(&snippet)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return false, ErrNotFound
        }
        return false, err
    }

    // Toggle the value
    newValue := !snippet.IsFavorite
    _, err = r.collection.UpdateOne(
        ctx,
        bson.M{"_id": id},
        bson.M{"$set": bson.M{"is_favorite": newValue, "updated_at": time.Now()}},
    )
    if err != nil {
        return false, err
    }

    return newValue, nil
}
```

## MongoDB vs PostgreSQL in DevJournal

| Feature | PostgreSQL (Journal) | MongoDB (Snippets) |
|---------|---------------------|-------------------|
| Schema | Fixed columns | Flexible documents |
| Metadata | Limited | Dynamic `metadata` map |
| Relations | Foreign keys | Embedded or referenced |
| Transactions | ACID | Multi-doc (less common) |
| Search | Full-text index | Text index with weights |
| Aggregation | SQL GROUP BY | Pipeline stages |
| Arrays | TEXT[] type | Native arrays |

## Key MongoDB Operations

### 1. BSON Query Operators

```go
// Equality
bson.M{"user_id": userID}

// Comparison
bson.M{"view_count": bson.M{"$gt": 10}}

// Array contains
bson.M{"tags": "angular"}  // Tags contains "angular"

// Array contains all
bson.M{"tags": bson.M{"$all": []string{"angular", "signals"}}}

// Logical OR
bson.M{"$or": []bson.M{
    {"language": "typescript"},
    {"language": "javascript"},
}}

// Text search
bson.M{"$text": bson.M{"$search": "signal store"}}
```

### 2. Update Operators

```go
// Set fields
bson.M{"$set": bson.M{"title": "New Title"}}

// Increment
bson.M{"$inc": bson.M{"view_count": 1}}

// Add to array (unique)
bson.M{"$addToSet": bson.M{"tags": "new-tag"}}

// Remove from array
bson.M{"$pull": bson.M{"tags": "old-tag"}}

// Unset field
bson.M{"$unset": bson.M{"metadata.deprecated": ""}}
```

### 3. Aggregation Pipeline

```go
// Language statistics with aggregation
pipeline := mongo.Pipeline{
    // Stage 1: Filter
    {{Key: "$match", Value: bson.M{"user_id": userID}}},

    // Stage 2: Group
    {{Key: "$group", Value: bson.M{
        "_id":           "$language",
        "count":         bson.M{"$sum": 1},
        "totalViews":    bson.M{"$sum": "$view_count"},
        "latestSnippet": bson.M{"$max": "$created_at"},
    }}},

    // Stage 3: Sort
    {{Key: "$sort", Value: bson.M{"count": -1}}},

    // Stage 4: Limit
    {{Key: "$limit", Value: 10}},
}
```

## Text Search with Weights

```go
// Create weighted text index
{
    Keys: bson.D{
        {Key: "title", Value: "text"},
        {Key: "description", Value: "text"},
        {Key: "code", Value: "text"},
    },
    Options: options.Index().SetWeights(bson.D{
        {Key: "title", Value: 10},       // Most important
        {Key: "description", Value: 5},  // Medium importance
        {Key: "code", Value: 1},         // Least important
    }),
}

// Search with score
findOptions := options.Find().
    SetProjection(bson.M{
        "score": bson.M{"$meta": "textScore"},
    }).
    SetSort(bson.D{
        {Key: "score", Value: bson.M{"$meta": "textScore"}},
    })

cursor, _ := collection.Find(ctx, bson.M{
    "$text": bson.M{"$search": "angular signals"},
}, findOptions)
```

## Connection String Format

```
mongodb://username:password@host:port/database?authSource=admin

# With replica set
mongodb://user:pass@host1:27017,host2:27017,host3:27017/dbname?replicaSet=rs0

# Example for Docker
mongodb://devjournal:devpass@localhost:27017/devjournal?authSource=admin
```

## Docker Setup

```yaml
# docker/docker-compose.yml
services:
  mongodb:
    image: mongo:7
    environment:
      MONGO_INITDB_ROOT_USERNAME: devjournal
      MONGO_INITDB_ROOT_PASSWORD: devpass
      MONGO_INITDB_DATABASE: devjournal
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
    healthcheck:
      test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  mongodb_data:
```

## Best Practices

1. **Use indexes wisely** - Create indexes for frequently queried fields
2. **Atomic operations** - Use `$inc`, `$set`, `$push` for atomic updates
3. **Projection** - Only fetch fields you need
4. **Pagination** - Always use `skip` and `limit` for large collections
5. **ObjectID vs UUID** - Use ObjectID for MongoDB-native IDs, UUID for cross-db references
6. **Schema validation** - Consider JSON Schema validation for important fields
7. **Connection pooling** - Configure pool size based on workload

## Key Takeaways

1. **Document model** - Ideal for flexible, nested data structures
2. **No joins** - Embed related data or reference with IDs
3. **Text search** - Built-in with weighted scoring
4. **Aggregation pipeline** - Powerful for analytics
5. **BSON types** - Use appropriate types (ObjectId, ISODate, etc.)
6. **Indexes** - Critical for query performance

## Next Steps

- [Angular Signal Store](./05-signal-store.md) - Frontend state for snippets
- [REST API Implementation](./02-rest-api.md) - Snippet API endpoints
