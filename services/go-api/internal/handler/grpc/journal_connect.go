package grpc

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"devjournal/internal/domain"
	"devjournal/internal/service"
	pb "devjournal/proto/devjournal/v1"
	"devjournal/proto/devjournal/v1/devjournalv1connect"
)

// JournalConnectHandler implements the Connect RPC JournalService
type JournalConnectHandler struct {
	devjournalv1connect.UnimplementedJournalServiceHandler
	journalService *service.JournalService
}

// NewJournalConnectHandler creates a new Connect RPC journal handler
func NewJournalConnectHandler(journalService *service.JournalService) *JournalConnectHandler {
	return &JournalConnectHandler{journalService: journalService}
}

// CreateEntry creates a new journal entry
func (h *JournalConnectHandler) CreateEntry(
	ctx context.Context,
	req *connect.Request[pb.CreateEntryRequest],
) (*connect.Response[pb.JournalEntry], error) {
	// Get user ID from context (set by auth interceptor)
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	// Create the entry
	domainReq := &domain.CreateJournalEntryRequest{
		Title:   req.Msg.Title,
		Content: req.Msg.Content,
		Mood:    req.Msg.Mood,
		Tags:    req.Msg.Tags,
	}

	entry, err := h.journalService.Create(ctx, userID, domainReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(domainToProtoJournalEntry(entry)), nil
}

// GetEntry retrieves a single journal entry by ID
func (h *JournalConnectHandler) GetEntry(
	ctx context.Context,
	req *connect.Request[pb.GetEntryRequest],
) (*connect.Response[pb.JournalEntry], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	entryID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	entry, err := h.journalService.GetByID(ctx, entryID, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if entry == nil {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}

	return connect.NewResponse(domainToProtoJournalEntry(entry)), nil
}

// ListEntries retrieves a paginated list of journal entries
func (h *JournalConnectHandler) ListEntries(
	ctx context.Context,
	req *connect.Request[pb.ListEntriesRequest],
) (*connect.Response[pb.ListEntriesResponse], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	var entries []domain.JournalEntry
	var total int

	if req.Msg.Mood != "" {
		entries, err = h.journalService.ListByMood(ctx, userID, req.Msg.Mood, int(req.Msg.Limit), int(req.Msg.Offset))
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		total = len(entries) // For mood filter, we don't have exact total
	} else {
		entries, total, err = h.journalService.List(ctx, userID, int(req.Msg.Limit), int(req.Msg.Offset))
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	protoEntries := make([]*pb.JournalEntry, len(entries))
	for i, entry := range entries {
		e := entry // Create a copy to avoid pointer issues
		protoEntries[i] = domainToProtoJournalEntry(&e)
	}

	return connect.NewResponse(&pb.ListEntriesResponse{
		Entries:    protoEntries,
		TotalCount: int32(total),
	}), nil
}

// UpdateEntry updates an existing journal entry
func (h *JournalConnectHandler) UpdateEntry(
	ctx context.Context,
	req *connect.Request[pb.UpdateEntryRequest],
) (*connect.Response[pb.JournalEntry], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	entryID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	domainReq := &domain.UpdateJournalEntryRequest{
		Title:   req.Msg.Title,
		Content: req.Msg.Content,
		Mood:    req.Msg.Mood,
		Tags:    req.Msg.Tags,
	}

	entry, err := h.journalService.Update(ctx, entryID, userID, domainReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(domainToProtoJournalEntry(entry)), nil
}

// DeleteEntry removes a journal entry
func (h *JournalConnectHandler) DeleteEntry(
	ctx context.Context,
	req *connect.Request[pb.DeleteEntryRequest],
) (*connect.Response[pb.DeleteEntryResponse], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	entryID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	err = h.journalService.Delete(ctx, entryID, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&pb.DeleteEntryResponse{Success: true}), nil
}

// SearchEntries searches journal entries by title or content
func (h *JournalConnectHandler) SearchEntries(
	ctx context.Context,
	req *connect.Request[pb.SearchEntriesRequest],
) (*connect.Response[pb.ListEntriesResponse], error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	entries, err := h.journalService.Search(ctx, userID, req.Msg.Query, int(req.Msg.Limit), int(req.Msg.Offset))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoEntries := make([]*pb.JournalEntry, len(entries))
	for i, entry := range entries {
		e := entry
		protoEntries[i] = domainToProtoJournalEntry(&e)
	}

	return connect.NewResponse(&pb.ListEntriesResponse{
		Entries:    protoEntries,
		TotalCount: int32(len(entries)),
	}), nil
}

// domainToProtoJournalEntry converts a domain JournalEntry to proto
func domainToProtoJournalEntry(entry *domain.JournalEntry) *pb.JournalEntry {
	return &pb.JournalEntry{
		Id:        entry.ID.String(),
		UserId:    entry.UserID.String(),
		Title:     entry.Title,
		Content:   entry.Content,
		Mood:      entry.Mood,
		Tags:      entry.Tags,
		CreatedAt: timestamppb.New(entry.CreatedAt),
		UpdatedAt: timestamppb.New(entry.UpdatedAt),
	}
}
