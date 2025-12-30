package mongodb

import (
	"context"
	"fmt"
	"time"

	"devjournal/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SnippetRepository handles snippet data persistence in MongoDB
type SnippetRepository struct {
	collection *mongo.Collection
}

// NewSnippetRepository creates a new snippet repository
func NewSnippetRepository(client *mongo.Client, dbName string) *SnippetRepository {
	collection := client.Database(dbName).Collection("snippets")

	// Create indexes for better query performance
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "tags", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "prog_lang", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "title", Value: "text"},
				{Key: "description", Value: "text"},
				{Key: "code", Value: "text"},
			},
		},
	}

	collection.Indexes().CreateMany(ctx, indexes)

	return &SnippetRepository{collection: collection}
}

// snippetDoc is the MongoDB document representation
// Note: Language uses "prog_lang" BSON tag to avoid conflict with MongoDB's
// reserved "language" field used for text index language override
type snippetDoc struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty"`
	UserID      string                 `bson:"user_id"`
	Title       string                 `bson:"title"`
	Description string                 `bson:"description"`
	Code        string                 `bson:"code"`
	Language    string                 `bson:"prog_lang"`
	Tags        []string               `bson:"tags"`
	Metadata    map[string]interface{} `bson:"metadata"`
	IsPublic    bool                   `bson:"is_public"`
	ViewsCount  int                    `bson:"views_count"`
	CreatedAt   time.Time              `bson:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at"`
}

func toDoc(s *domain.Snippet) *snippetDoc {
	doc := &snippetDoc{
		UserID:      s.UserID,
		Title:       s.Title,
		Description: s.Description,
		Code:        s.Code,
		Language:    s.Language,
		Tags:        s.Tags,
		Metadata:    s.Metadata,
		IsPublic:    s.IsPublic,
		ViewsCount:  s.ViewsCount,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
	if s.ID != "" {
		if oid, err := primitive.ObjectIDFromHex(s.ID); err == nil {
			doc.ID = oid
		}
	}
	return doc
}

func fromDoc(doc *snippetDoc) *domain.Snippet {
	return &domain.Snippet{
		ID:          doc.ID.Hex(),
		UserID:      doc.UserID,
		Title:       doc.Title,
		Description: doc.Description,
		Code:        doc.Code,
		Language:    doc.Language,
		Tags:        doc.Tags,
		Metadata:    doc.Metadata,
		IsPublic:    doc.IsPublic,
		ViewsCount:  doc.ViewsCount,
		CreatedAt:   doc.CreatedAt,
		UpdatedAt:   doc.UpdatedAt,
	}
}

// Create inserts a new snippet
func (r *SnippetRepository) Create(ctx context.Context, snippet *domain.Snippet) error {
	snippet.CreatedAt = time.Now().UTC()
	snippet.UpdatedAt = snippet.CreatedAt

	doc := toDoc(snippet)
	doc.ID = primitive.NewObjectID()

	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to create snippet: %w", err)
	}

	// Set the ID back to the snippet
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		snippet.ID = oid.Hex()
	}
	return nil
}

// FindByID retrieves a snippet by ID
func (r *SnippetRepository) FindByID(ctx context.Context, id string) (*domain.Snippet, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, nil // Invalid ID format
	}
	filter := bson.M{"_id": oid}

	var doc snippetDoc
	err = r.collection.FindOne(ctx, filter).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find snippet: %w", err)
	}
	return fromDoc(&doc), nil
}

// FindByUserID retrieves all snippets for a user with pagination
func (r *SnippetRepository) FindByUserID(ctx context.Context, userID string, limit, offset int64) ([]domain.Snippet, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find snippets: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []snippetDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("failed to decode snippets: %w", err)
	}

	snippets := make([]domain.Snippet, len(docs))
	for i, doc := range docs {
		snippets[i] = *fromDoc(&doc)
	}
	return snippets, nil
}

// FindByTags retrieves snippets matching any of the given tags
func (r *SnippetRepository) FindByTags(ctx context.Context, userID string, tags []string, limit, offset int64) ([]domain.Snippet, error) {
	filter := bson.M{
		"user_id": userID,
		"tags":    bson.M{"$in": tags},
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find snippets by tags: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []snippetDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("failed to decode snippets: %w", err)
	}

	snippets := make([]domain.Snippet, len(docs))
	for i, doc := range docs {
		snippets[i] = *fromDoc(&doc)
	}
	return snippets, nil
}

// FindByLanguage retrieves snippets by programming language
func (r *SnippetRepository) FindByLanguage(ctx context.Context, userID, language string, limit, offset int64) ([]domain.Snippet, error) {
	filter := bson.M{
		"user_id":   userID,
		"prog_lang": language,
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find snippets by language: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []snippetDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("failed to decode snippets: %w", err)
	}

	snippets := make([]domain.Snippet, len(docs))
	for i, doc := range docs {
		snippets[i] = *fromDoc(&doc)
	}
	return snippets, nil
}

// Search performs full-text search on snippets
func (r *SnippetRepository) Search(ctx context.Context, userID, query string, limit, offset int64) ([]domain.Snippet, error) {
	filter := bson.M{
		"user_id": userID,
		"$text":   bson.M{"$search": query},
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "score", Value: bson.M{"$meta": "textScore"}}}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search snippets: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []snippetDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("failed to decode snippets: %w", err)
	}

	snippets := make([]domain.Snippet, len(docs))
	for i, doc := range docs {
		snippets[i] = *fromDoc(&doc)
	}
	return snippets, nil
}

// Update updates an existing snippet
func (r *SnippetRepository) Update(ctx context.Context, snippet *domain.Snippet) error {
	oid, err := primitive.ObjectIDFromHex(snippet.ID)
	if err != nil {
		return fmt.Errorf("invalid snippet ID: %w", err)
	}

	snippet.UpdatedAt = time.Now().UTC()

	filter := bson.M{"_id": oid, "user_id": snippet.UserID}
	update := bson.M{"$set": bson.M{
		"title":       snippet.Title,
		"description": snippet.Description,
		"code":        snippet.Code,
		"prog_lang":   snippet.Language,
		"tags":        snippet.Tags,
		"metadata":    snippet.Metadata,
		"is_public":   snippet.IsPublic,
		"updated_at":  snippet.UpdatedAt,
	}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update snippet: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("snippet not found or unauthorized")
	}
	return nil
}

// Delete removes a snippet
func (r *SnippetRepository) Delete(ctx context.Context, id, userID string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid snippet ID: %w", err)
	}

	filter := bson.M{"_id": oid, "user_id": userID}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete snippet: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("snippet not found or unauthorized")
	}
	return nil
}

// IncrementViews increments the view count for a snippet
func (r *SnippetRepository) IncrementViews(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid snippet ID: %w", err)
	}

	filter := bson.M{"_id": oid}
	update := bson.M{"$inc": bson.M{"views_count": 1}}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to increment views: %w", err)
	}
	return nil
}

// Count returns the total number of snippets for a user
func (r *SnippetRepository) Count(ctx context.Context, userID string) (int64, error) {
	filter := bson.M{"user_id": userID}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count snippets: %w", err)
	}
	return count, nil
}

// GetLanguageStats returns snippet counts grouped by language
func (r *SnippetRepository) GetLanguageStats(ctx context.Context, userID string) (map[string]int64, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"user_id": userID}},
		{"$group": bson.M{
			"_id":   "$prog_lang",
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get language stats: %w", err)
	}
	defer cursor.Close(ctx)

	stats := make(map[string]int64)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode stats: %w", err)
		}
		stats[result.ID] = result.Count
	}

	return stats, nil
}
