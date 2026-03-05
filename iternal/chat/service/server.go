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