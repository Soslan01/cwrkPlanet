package room

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cwrk-planet/api-gateway/pkg/errs"
	"github.com/cwrk-planet/api-gateway/pkg/httputil"

	roomv1 "github.com/cwrk-planet/room-service/proto/gen/room/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Client — API для HTTP-слоя gateway.
type Client interface {
	CreateRoom(ctx context.Context, authHeader string, userID int64, in CreateRoomRequest) (RoomItem, error)
	ListRooms(ctx context.Context, authHeader string, userID int64, limit int64, cursor string) (RoomsListResponse, error)
	GetRoom(ctx context.Context, authHeader string, userID int64, id string) (RoomItem, error)
	Join(ctx context.Context, authHeader string, userID int64, id string) (JoinRoomResponse, error)
	Leave(ctx context.Context, authHeader string, userID int64, id string) error
	Participants(ctx context.Context, authHeader string, userID int64, id string) (ParticipantsResponse, error)
	ChatHistory(ctx context.Context, authHeader string, userID int64, roomID string, after string, limit int32) (ChatHistoryResponse, error)
	Close() error
}

type client struct {
	conn    *grpc.ClientConn
	room    roomv1.RoomServiceClient
	timeout time.Duration
}

type Options struct {
	Target      string
	Timeout     time.Duration
	UseInsecure bool // зарезервировано под TLS в "Когда-нибудь будет" |:)
}

func New(opts Options) (Client, error) {
	if opts.Target == "" {
		return nil, fmt.Errorf("room client: empty target")
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 5 * time.Second
	}
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(opts.Target, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("room client: new client failed: %w", err)
	}

	return &client{
		conn:    conn,
		room:    roomv1.NewRoomServiceClient(conn),
		timeout: opts.Timeout,
	}, nil
}

func (c *client) Close() error { return c.conn.Close() }

// withOutboundMeta — добавляет x-request-id, authorization, x-user-id в metadata.
func withOutboundMeta(ctx context.Context, authHeader string, userID int64) context.Context {
	// X-Request-ID
	if rid, ok := httputil.FromContext(ctx); ok && rid != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", rid)
	}
	// Authorization
	if authHeader != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authHeader)
	}
	// X-User-ID
	if userID != 0 {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-user-id", strconv.FormatInt(userID, 10))
	}
	return ctx
}

func (c *client) CreateRoom(ctx context.Context, authHeader string, userID int64, in CreateRoomRequest) (RoomItem, error) {
	rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	rpcCtx = withOutboundMeta(rpcCtx, authHeader, userID)

	req := &roomv1.CreateRoomRequest{
		Name: in.Name,
		Max:  in.Max,
	}
	res, err := c.room.CreateRoom(rpcCtx, req)
	if err != nil {
		return RoomItem{}, fmt.Errorf("%w: %v", errs.ErrUpstream, err)
	}

	return mapRoom(res.GetRoom()), nil
}

func (c *client) ListRooms(ctx context.Context, authHeader string, userID int64, limit int64, cursor string) (RoomsListResponse, error) {
	rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	rpcCtx = withOutboundMeta(rpcCtx, authHeader, userID)

	req := &roomv1.ListRoomsRequest{Limit: int32(limit), Cursor: cursor}
	res, err := c.room.ListRooms(rpcCtx, req)
	if err != nil {
		return RoomsListResponse{}, fmt.Errorf("%w: %v", errs.ErrUpstream, err)
	}

	out := RoomsListResponse{
		Items:      make([]RoomItem, 0, len(res.GetItems())),
		NextCursor: res.GetNextCursor(),
	}
	for _, r := range res.GetItems() {
		out.Items = append(out.Items, mapRoom(r))
	}

	return out, nil
}

func (c *client) GetRoom(ctx context.Context, authHeader string, userID int64, id string) (RoomItem, error) {
	rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	rpcCtx = withOutboundMeta(rpcCtx, authHeader, userID)

	res, err := c.room.GetRoom(rpcCtx, &roomv1.GetRoomRequest{Id: id})
	if err != nil {
		return RoomItem{}, fmt.Errorf("%w: %v", errs.ErrUpstream, err)
	}

	return mapRoom(res.GetRoom()), nil
}

func (c *client) Join(ctx context.Context, authHeader string, userID int64, id string) (JoinRoomResponse, error) {
	rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	rpcCtx = withOutboundMeta(rpcCtx, authHeader, userID)

	res, err := c.room.JoinRoom(rpcCtx, &roomv1.JoinRoomRequest{Id: id})
	if err != nil {
		return JoinRoomResponse{}, fmt.Errorf("%w: %v", errs.ErrUpstream, err)
	}

	return mapJoin(res), nil
}

func (c *client) Leave(ctx context.Context, authHeader string, userID int64, id string) error {
	rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	rpcCtx = withOutboundMeta(rpcCtx, authHeader, userID)

	_, err := c.room.LeaveRoom(rpcCtx, &roomv1.LeaveRoomRequest{Id: id})
	if err != nil {
		return fmt.Errorf("%w: %v", errs.ErrUpstream, err)
	}

	return nil
}

func (c *client) Participants(ctx context.Context, authHeader string, userID int64, id string) (ParticipantsResponse, error) {
	rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	rpcCtx = withOutboundMeta(rpcCtx, authHeader, userID)

	res, err := c.room.ListParticipants(rpcCtx, &roomv1.ListParticipantsRequest{Id: id})
	if err != nil {
		return ParticipantsResponse{}, fmt.Errorf("%w: %v", errs.ErrUpstream, err)
	}
	out := ParticipantsResponse{Items: make([]ParticipantItem, 0, len(res.GetItems()))}
	for _, p := range res.GetItems() {
		item := ParticipantItem{
			UserID:   p.GetUserId(),
			JoinedAt: p.GetJoinedAt().AsTime(),
			LastSeen: p.GetLastSeen().AsTime(),
		}
		out.Items = append(out.Items, item)
	}

	return out, nil
}

func (c *client) ChatHistory(ctx context.Context, authHeader string, userID int64, roomID string, after string, limit int32) (ChatHistoryResponse, error) {
	rpcCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	rpcCtx = withOutboundMeta(rpcCtx, authHeader, userID)

	req := &roomv1.GetChatHistoryRequest{
		Id:    roomID,
		After: after,
		Limit: limit,
	}
	res, err := c.room.GetChatHistory(rpcCtx, req)
	if err != nil {
		return ChatHistoryResponse{}, fmt.Errorf("%w: %v", errs.ErrUpstream, err)
	}
	out := ChatHistoryResponse{
		Items:      make([]ChatMessageItem, 0, len(res.GetItems())),
		NextCursor: res.GetNextCursor(),
	}
	for _, m := range res.GetItems() {
		item := ChatMessageItem{
			ID:     m.GetId(),
			RoomID: m.GetRoomId(),
			UserID: m.GetUserId(),
			Text:   m.GetText(),
		}
		if ts := m.GetCreatedAt(); ts != nil {
			item.CreatedAt = ts.AsTime()
		}
		if rt := m.GetReplyTo(); strings.TrimSpace(rt) != "" {
			item.ReplyTo = rt
		}
		out.Items = append(out.Items, item)
	}

	return out, nil
}

func mapRoom(in *roomv1.Room) RoomItem {
	if in == nil {
		return RoomItem{}
	}
	out := RoomItem{
		ID:              in.GetId(),
		Name:            in.GetName(),
		MaxParticipants: in.GetMaxParticipants(),
	}
	if ts := in.GetCreatedAt(); ts != nil {
		out.CreatedAt = ts.AsTime()
	}

	return out
}

func mapJoin(in *roomv1.JoinRoomResponse) JoinRoomResponse {
	var out JoinRoomResponse
	if in == nil {
		return out
	}
	out.RoomID = in.GetRoomId()
	out.PeerID = in.GetPeerId()

	return out
}
