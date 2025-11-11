package http

import (
	"net/http"
	"time"

	"github.com/cwrk-planet/room-service/internal/service"
	httpmw "github.com/cwrk-planet/room-service/internal/transport/http/middleware"
	"github.com/cwrk-planet/room-service/internal/transport/ws"

	"github.com/go-chi/chi/v5"
	middlewareChi "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler, memberSvc *service.MemberService, wsServer *ws.Server) http.Handler {
	r := chi.NewRouter()
	r.Use(middlewareChi.RequestID)
	r.Use(middlewareChi.RealIP)
	r.Use(middlewareChi.Recoverer)

	// WS endpoint
	r.Get("/ws/rooms/{id}", wsServer.HandleWS)

	// Все маршруты требуют access_token и user_id
	r.Group(func(pr chi.Router) {
		pr.Use(httpmw.AuthMiddleware)
		pr.Use(httpmw.HeartbeatMiddleware(memberSvc))
		pr.Use(middlewareChi.Timeout(30 * time.Second))

		pr.Route("/rooms", func(rm chi.Router) {
			rm.Post("/", h.CreateRoom)
			rm.Get("/", h.ListRooms)

			rm.Route("/{id}", func(rr chi.Router) {
				rr.Get("/", h.GetRoom)
				rr.Post("/join", h.JoinRoom)
				rr.Post("/leave", h.LeaveRoom)
				rr.Get("/participants", h.GetParticipants)
				rr.Get("/chat", h.GetChatHistory)
			})
		})
	})

	// health
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return r
}
