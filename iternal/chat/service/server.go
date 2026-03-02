package chat

import (
	"context"

	"github.com/chimort/course_project2/api/proto/chatpb"
)

type ChatServer struct {
	chatpb.UnimplementedChatServiceServer
	service *ChatService
}

func NewChatServer(service *ChatService) *ChatServer {
	return &ChatServer{
		service: service,
	}
}

func (s *ChatServer) CreateChat(
	ctx context.Context,
	req *chatpb.CreateChatRequest,
) (*chatpb.CreateChatResponse, error) {

	if req.User1 == "" || req.User2 == "" {
		return &chatpb.CreateChatResponse{Ok: false}, nil
	}

	chat, err := s.service.CreateChat(ctx, req.User1, req.User2)
	if err != nil {
		return nil, err
	}

	return &chatpb.CreateChatResponse{
		Ok:     true,
		ChatId: chat.ID,
	}, nil
}