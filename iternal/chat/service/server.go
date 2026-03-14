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
		})
	}

	return &chatpb.GetUserChatsResponse{
		Chats: resp,
	}, nil
}
