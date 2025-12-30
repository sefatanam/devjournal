package grpc

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"devjournal/internal/domain"
	"devjournal/internal/service"
	pb "devjournal/proto/devjournal/v1"
	"devjournal/proto/devjournal/v1/devjournalv1connect"
)

// SnippetConnectHandler implements the Connect RPC SnippetService
type SnippetConnectHandler struct {
	devjournalv1connect.UnimplementedSnippetServiceHandler
	snippetService *service.SnippetService
}

// NewSnippetConnectHandler creates a new Connect RPC snippet handler
func NewSnippetConnectHandler(snippetService *service.SnippetService) *SnippetConnectHandler {
	return &SnippetConnectHandler{snippetService: snippetService}
}

// CreateSnippet creates a new code snippet
func (h *SnippetConnectHandler) CreateSnippet(
	ctx context.Context,
	req *connect.Request[pb.CreateSnippetRequest],
) (*connect.Response[pb.Snippet], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	metadata := structToMap(req.Msg.Metadata)

	domainReq := &domain.CreateSnippetRequest{
		Title:       req.Msg.Title,
		Description: req.Msg.Description,
		Code:        req.Msg.Code,
		Language:    req.Msg.Language,
		Tags:        req.Msg.Tags,
		Metadata:    metadata,
		IsPublic:    req.Msg.IsPublic,
	}

	snippet, err := h.snippetService.Create(ctx, userID.String(), domainReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(domainToProtoSnippet(snippet)), nil
}

// GetSnippet retrieves a single snippet by ID
func (h *SnippetConnectHandler) GetSnippet(
	ctx context.Context,
	req *connect.Request[pb.GetSnippetRequest],
) (*connect.Response[pb.Snippet], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	snippet, err := h.snippetService.GetByID(ctx, req.Msg.Id, userID.String())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if snippet == nil {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}

	return connect.NewResponse(domainToProtoSnippet(snippet)), nil
}

// ListSnippets retrieves a paginated list of snippets
func (h *SnippetConnectHandler) ListSnippets(
	ctx context.Context,
	req *connect.Request[pb.ListSnippetsRequest],
) (*connect.Response[pb.ListSnippetsResponse], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	var snippets []domain.Snippet
	var total int64

	if req.Msg.Language != "" {
		snippets, err = h.snippetService.ListByLanguage(ctx, userID.String(), req.Msg.Language, int64(req.Msg.Limit), int64(req.Msg.Offset))
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		total = int64(len(snippets))
	} else if len(req.Msg.Tags) > 0 {
		snippets, err = h.snippetService.ListByTags(ctx, userID.String(), req.Msg.Tags, int64(req.Msg.Limit), int64(req.Msg.Offset))
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		total = int64(len(snippets))
	} else {
		snippets, total, err = h.snippetService.List(ctx, userID.String(), int64(req.Msg.Limit), int64(req.Msg.Offset))
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	protoSnippets := make([]*pb.Snippet, len(snippets))
	for i, snippet := range snippets {
		s := snippet
		protoSnippets[i] = domainToProtoSnippet(&s)
	}

	return connect.NewResponse(&pb.ListSnippetsResponse{
		Snippets:   protoSnippets,
		TotalCount: total,
	}), nil
}

// UpdateSnippet updates an existing snippet
func (h *SnippetConnectHandler) UpdateSnippet(
	ctx context.Context,
	req *connect.Request[pb.UpdateSnippetRequest],
) (*connect.Response[pb.Snippet], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	metadata := structToMap(req.Msg.Metadata)

	domainReq := &domain.UpdateSnippetRequest{
		Title:       req.Msg.Title,
		Description: req.Msg.Description,
		Code:        req.Msg.Code,
		Language:    req.Msg.Language,
		Tags:        req.Msg.Tags,
		Metadata:    metadata,
		IsPublic:    req.Msg.IsPublic,
	}

	snippet, err := h.snippetService.Update(ctx, req.Msg.Id, userID.String(), domainReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(domainToProtoSnippet(snippet)), nil
}

// DeleteSnippet removes a snippet
func (h *SnippetConnectHandler) DeleteSnippet(
	ctx context.Context,
	req *connect.Request[pb.DeleteSnippetRequest],
) (*connect.Response[pb.DeleteSnippetResponse], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	err = h.snippetService.Delete(ctx, req.Msg.Id, userID.String())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&pb.DeleteSnippetResponse{Success: true}), nil
}

// SearchSnippets performs full-text search on snippets
func (h *SnippetConnectHandler) SearchSnippets(
	ctx context.Context,
	req *connect.Request[pb.SearchSnippetsRequest],
) (*connect.Response[pb.ListSnippetsResponse], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	snippets, err := h.snippetService.Search(ctx, userID.String(), req.Msg.Query, int64(req.Msg.Limit), int64(req.Msg.Offset))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoSnippets := make([]*pb.Snippet, len(snippets))
	for i, snippet := range snippets {
		s := snippet
		protoSnippets[i] = domainToProtoSnippet(&s)
	}

	return connect.NewResponse(&pb.ListSnippetsResponse{
		Snippets:   protoSnippets,
		TotalCount: int64(len(snippets)),
	}), nil
}

// GetLanguageStats returns snippet counts grouped by language
func (h *SnippetConnectHandler) GetLanguageStats(
	ctx context.Context,
	req *connect.Request[pb.GetLanguageStatsRequest],
) (*connect.Response[pb.GetLanguageStatsResponse], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	stats, err := h.snippetService.GetLanguageStats(ctx, userID.String())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&pb.GetLanguageStatsResponse{
		LanguageCounts: stats,
	}), nil
}

// domainToProtoSnippet converts a domain Snippet to proto
func domainToProtoSnippet(snippet *domain.Snippet) *pb.Snippet {
	metadata, _ := structpb.NewStruct(snippet.Metadata)

	return &pb.Snippet{
		Id:          snippet.ID,
		UserId:      snippet.UserID,
		Title:       snippet.Title,
		Description: snippet.Description,
		Code:        snippet.Code,
		Language:    snippet.Language,
		Tags:        snippet.Tags,
		Metadata:    metadata,
		IsPublic:    snippet.IsPublic,
		ViewsCount:  int32(snippet.ViewsCount),
		CreatedAt:   timestamppb.New(snippet.CreatedAt),
		UpdatedAt:   timestamppb.New(snippet.UpdatedAt),
	}
}

// structToMap converts a protobuf Struct to a Go map
func structToMap(s *structpb.Struct) map[string]interface{} {
	if s == nil {
		return nil
	}
	return s.AsMap()
}
