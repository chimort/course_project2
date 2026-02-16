package matching

import (
	"context"

	"github.com/chimort/course_project2/api/proto/matchingpb"
)

type MatchingServer struct {
	matchingpb.UnimplementedMatchingServiceServer
	service *MatchingService
}

func NewMatchingServer(service *MatchingService) *MatchingServer {
	return &MatchingServer{service: service}
}

func (s *MatchingServer) JoinQueue(ctx context.Context, req *matchingpb.JoinQueueRequest) (*matchingpb.JoinQueueResponse, error) {
	return s.service.JoinQueue(ctx, req)
}

func (s *MatchingServer) LeaveQueue(ctx context.Context, req *matchingpb.LeaveQueueRequest) (*matchingpb.LeaveQueueResponse, error) {
	return s.service.LeaveQueue(ctx, req)
}

func (s *MatchingServer) ListQueue(ctx context.Context, req *matchingpb.ListQueueRequest) (*matchingpb.ListQueueResponse, error) {
	return s.service.ListQueue(ctx, req)
}
