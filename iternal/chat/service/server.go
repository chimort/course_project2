package chat

import (
	"context"
	"strconv"

	chatpb "github.com/chimort/course_project2/api/proto/chatpb"
)

type ChatServer struct {
	chatpb.UnimplementedChatServiceServer
	service *ChatService
}

func NewChatServer(service *ChatService) *ChatServer {
	return &ChatServer{service: service}
}

func (s *ChatServer) CreateChat(
	ctx context.Context,
	req *chatpb.CreateChatRequest,
) (*chatpb.CreateChatResponse, error) {

	chat, err := s.service.CreateChat(
		ctx,
		req.User1,
		req.User2,
	)

	if err != nil {
		return nil, err
	}

	return &chatpb.CreateChatResponse{
		ChatId: chat.ID,
	}, nil
}

func (s *ChatServer) SendMessage(
	ctx context.Context,
	req *chatpb.SendMessageRequest,
) (*chatpb.SendMessageResponse, error) {

	chatID, _ := strconv.Atoi(req.ChatId)

	err := s.service.SendMessage(
		ctx,
		chatID,
		req.Sender,
		req.Content,
	)

	if err != nil {
		return nil, err
	}

	return &chatpb.SendMessageResponse{
		Ok: true,
	}, nil
}

func (s *ChatServer) GetParticipants(
	ctx context.Context,
	req *chatpb.GetParticipantsRequest,
) (*chatpb.GetParticipantsResponse, error) {

	chatID, _ := strconv.Atoi(req.ChatId)

	users, err := s.service.GetParticipants(
		ctx,
		chatID,
	)

	if err != nil {
		return nil, err
	}

	return &chatpb.GetParticipantsResponse{
		Usernames: users,
	}, nil
}

func (s *ChatServer) GetMessages(
	ctx context.Context,
	req *chatpb.GetMessagesRequest,
) (*chatpb.GetMessagesResponse, error) {

	chatID, _ := strconv.Atoi(req.ChatId)

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 100
	}

	messages, err := s.service.GetMessages(ctx, chatID, limit)
	if err != nil {
		return nil, err
	}

	resp := make([]*chatpb.ChatMessage, 0, len(messages))
	for _, m := range messages {
		resp = append(resp, &chatpb.ChatMessage{
			Sender:    m.Sender,
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
		})
	}

	return &chatpb.GetMessagesResponse{
		Messages: resp,
	}, nil
}

func (s *ChatServer) GetUserChats(
	ctx context.Context,
	req *chatpb.GetUserChatsRequest,
) (*chatpb.GetUserChatsResponse, error) {

	chats, err := s.service.GetUserChats(ctx, req.Username)
	if req.SortBy != "" || req.FilterMode != "" || req.FilterTag != "" || req.PeerQuery != "" {
		chats, err = s.service.GetUserChatsFiltered(ctx, req.Username, req.SortBy, req.FilterMode, req.FilterTag, req.PeerQuery)
	}
	if err != nil {
		return nil, err
	}

	resp := make([]*chatpb.ChatPreview, 0, len(chats))
	for _, chat := range chats {
		resp = append(resp, &chatpb.ChatPreview{
			ChatId:        chat.ChatID,
			PeerUsername:  chat.PeerUsername,
			LastMessage:   chat.LastMessage,
			LastMessageAt: chat.LastMessageAt,
			HasUnread:     chat.HasUnread,
			MatchHint:     chat.MatchHint,
			MatchTags:     chat.MatchTags,
			SearchMode:    chat.SearchMode,
		})
	}

	return &chatpb.GetUserChatsResponse{
		Chats: resp,
	}, nil
}

func (s *ChatServer) MarkChatRead(
	ctx context.Context,
	req *chatpb.MarkChatReadRequest,
) (*chatpb.MarkChatReadResponse, error) {
	chatID, _ := strconv.Atoi(req.ChatId)

	if err := s.service.MarkChatRead(ctx, chatID, req.Username); err != nil {
		return nil, err
	}

	return &chatpb.MarkChatReadResponse{Ok: true}, nil
}

func (s *ChatServer) SetChatMatchTags(
	ctx context.Context,
	req *chatpb.SetChatMatchTagsRequest,
) (*chatpb.SetChatMatchTagsResponse, error) {
	chatID, _ := strconv.Atoi(req.ChatId)

	if req.SearchMode != "" {
		if err := s.service.SetChatMatchMetadata(ctx, chatID, req.Tags, req.SearchMode); err != nil {
			return nil, err
		}
	} else if err := s.service.SetChatMatchTags(ctx, chatID, req.Tags); err != nil {
		return nil, err
	}

	return &chatpb.SetChatMatchTagsResponse{Ok: true}, nil
}

func (s *ChatServer) BlockUser(
	ctx context.Context,
	req *chatpb.BlockUserRequest,
) (*chatpb.BlockUserResponse, error) {
	if err := s.service.BlockUser(ctx, req.BlockerUsername, req.BlockedUsername); err != nil {
		return nil, err
	}

	return &chatpb.BlockUserResponse{Ok: true}, nil
}

func (s *ChatServer) GetBlockStatus(
	ctx context.Context,
	req *chatpb.GetBlockStatusRequest,
) (*chatpb.GetBlockStatusResponse, error) {
	isBlocked, blockedByUser1, blockedByUser2, err := s.service.GetBlockStatus(ctx, req.User1, req.User2)
	if err != nil {
		return nil, err
	}

	return &chatpb.GetBlockStatusResponse{
		IsBlocked:      isBlocked,
		BlockedByUser1: blockedByUser1,
		BlockedByUser2: blockedByUser2,
	}, nil
}

func (s *ChatServer) UnblockUser(
	ctx context.Context,
	req *chatpb.UnblockUserRequest,
) (*chatpb.UnblockUserResponse, error) {
	if err := s.service.UnblockUser(ctx, req.BlockerUsername, req.BlockedUsername); err != nil {
		return nil, err
	}

	return &chatpb.UnblockUserResponse{Ok: true}, nil
}
