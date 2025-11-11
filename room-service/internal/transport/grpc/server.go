package grpcx

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/cwrk-planet/room-service/internal/domain"
	"github.com/cwrk-planet/room-service/internal/service"

	roomv1 "github.com/cwrk-planet/room-service/proto/gen/room/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	mdAuthorization = "authorization"
	mdUserID        = "x-user-id"
)

type Server struct {
	roomv1.UnimplementedRoomServiceServer

	roomSvc   *service.RoomService
	memberSvc *service.MemberService
	chatSvc   *service.ChatService
}

func NewServer(
	roomSvc *service.RoomService,
	memberSvc *service.MemberService,
	chatSvc *service.ChatService,
) *Server {
	return &Server{
		roomSvc:   roomSvc,
		memberSvc: memberSvc,
		chatSvc:   chatSvc,
	}
}

func Register(grpcServer *grpc.Server, s *Server) {
	roomv1.RegisterRoomServiceServer(grpcServer, s)
}

// -------- helpers --------

func userFromMD(ctx context.Context) (token, userID string, _ error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "", status.Error(codes.Unauthenticated, "missing metadata")
	}
	// Authorization: Bearer <access_token>
	auth := first(md.Get(mdAuthorization))
	if auth == "" {
		return "", "", status.Error(codes.Unauthenticated, "missing authorization")
	}
	if !strings.HasPrefix(strings.ToLower(auth), "bearer ") || len(auth) <= 7 {
		return "", "", status.Error(codes.Unauthenticated, "invalid authorization")
	}
	token = strings.TrimSpace(auth[7:])

	userID = first(md.Get(mdUserID))
	if userID == "" {
		return "", "", status.Error(codes.Unauthenticated, "missing x-user-id")
	}

	return token, userID, nil
}

func first(ss []string) string {
	if len(ss) == 0 {
		return ""
	}

	return ss[0]
}

func valueOrEmpty(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func mapRoom(r *domain.Room) *roomv1.Room {
	return &roomv1.Room{
		Id:              r.ID,
		Name:            r.Name,
		MaxParticipants: r.MaxParticipants,
		CreatedAt:       timestamppb.New(r.CreatedAt),
	}
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, domain.ErrRoomNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrRoomFull):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrAlreadyJoined):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrNotInRoom):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func mapChat(m domain.ChatMessage) *roomv1.ChatMessage {
	return &roomv1.ChatMessage{
		Id:        m.ID,
		RoomId:    m.RoomID,
		UserId:    strconv.FormatInt(m.UserID, 10),
		Text:      m.Text,
		CreatedAt: timestamppb.New(m.CreatedAt),
		ReplyTo:   valueOrEmpty(m.ReplyTo),
	}
}

// -------- methods --------

func (s *Server) CreateRoom(ctx context.Context, in *roomv1.CreateRoomRequest) (*roomv1.CreateRoomResponse, error) {
	if _, _, err := userFromMD(ctx); err != nil {
		return nil, err
	}
	room, err := s.roomSvc.CreateRoom(ctx, in.GetName(), in.GetMax())
	if err != nil {
		return nil, mapErr(err)
	}

	return &roomv1.CreateRoomResponse{Room: mapRoom(room)}, nil
}

func (s *Server) ListRooms(ctx context.Context, in *roomv1.ListRoomsRequest) (*roomv1.ListRoomsResponse, error) {
	if _, _, err := userFromMD(ctx); err != nil {
		return nil, err
	}
	limit := int(in.GetLimit())
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	items, cursor, err := s.roomSvc.ListRooms(ctx, limit, in.GetCursor())
	if err != nil {
		return nil, mapErr(err)
	}
	out := &roomv1.ListRoomsResponse{
		Items:      make([]*roomv1.Room, 0, len(items)),
		NextCursor: cursor,
	}
	for i := range items {
		out.Items = append(out.Items, mapRoom(&items[i]))
	}

	return out, nil
}

func (s *Server) GetRoom(ctx context.Context, in *roomv1.GetRoomRequest) (*roomv1.GetRoomResponse, error) {
	if _, _, err := userFromMD(ctx); err != nil {
		return nil, err
	}
	room, err := s.roomSvc.GetRoom(ctx, in.GetId())
	if err != nil {
		return nil, mapErr(err)
	}

	return &roomv1.GetRoomResponse{Room: mapRoom(room)}, nil
}

func (s *Server) JoinRoom(ctx context.Context, in *roomv1.JoinRoomRequest) (*roomv1.JoinRoomResponse, error) {
	_, userID, err := userFromMD(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid x-user-id")
	}

	if _, err := s.memberSvc.JoinRoom(ctx, in.GetId(), uid); err != nil && !errors.Is(err, domain.ErrAlreadyJoined) {
		return nil, mapErr(err)
	}

	return &roomv1.JoinRoomResponse{
		RoomId: in.GetId(),
		PeerId: userID,
	}, nil
}

func (s *Server) LeaveRoom(ctx context.Context, in *roomv1.LeaveRoomRequest) (*roomv1.LeaveRoomResponse, error) {
	_, userID, err := userFromMD(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid x-user-id")
	}
	if err := s.memberSvc.LeaveRoom(ctx, in.GetId(), uid); err != nil && !errors.Is(err, domain.ErrNotInRoom) {
		return nil, mapErr(err)
	}

	return &roomv1.LeaveRoomResponse{}, nil
}

func (s *Server) ListParticipants(ctx context.Context, in *roomv1.ListParticipantsRequest) (*roomv1.ListParticipantsResponse, error) {
	if _, _, err := userFromMD(ctx); err != nil {
		return nil, err
	}
	parts, err := s.memberSvc.ListParticipants(ctx, in.GetId())
	if err != nil {
		return nil, mapErr(err)
	}
	out := &roomv1.ListParticipantsResponse{
		Items: make([]*roomv1.Participant, 0, len(parts)),
	}
	for _, p := range parts {
		out.Items = append(out.Items, &roomv1.Participant{
			UserId:   strconv.FormatInt(p.UserID, 10),
			JoinedAt: timestamppb.New(p.JoinedAt),
			LastSeen: timestamppb.New(p.LastSeen),
		})
	}

	return out, nil
}

func (s *Server) GetChatHistory(ctx context.Context, in *roomv1.GetChatHistoryRequest) (*roomv1.GetChatHistoryResponse, error) {
	if _, _, err := userFromMD(ctx); err != nil {
		return nil, err
	}
	if s.chatSvc == nil {
		return nil, status.Error(codes.Unimplemented, "chat service disabled")
	}
	limit := int(in.GetLimit())
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	items, next, err := s.chatSvc.History(ctx, in.GetId(), in.GetAfter(), limit)
	if err != nil {
		return nil, mapErr(err)
	}

	out := &roomv1.GetChatHistoryResponse{
		Items:      make([]*roomv1.ChatMessage, 0, len(items)),
		NextCursor: next,
	}
	for _, m := range items {
		out.Items = append(out.Items, mapChat(m))
	}

	return out, nil
}
