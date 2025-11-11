package http

import (
	"net/http"
	"time"

	appauth "github.com/cwrk-planet/api-gateway/internal/app/auth"
	approom "github.com/cwrk-planet/api-gateway/internal/app/room"
	"github.com/cwrk-planet/api-gateway/pkg/httputil"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Deps struct {
	AuthClient appauth.Client
	RoomClient approom.Client
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(httputil.MiddlewareRequestID)
	r.Use(httputil.MiddlewareLogging)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID", "X-User-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// health
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httputil.OK(w, map[string]string{"status": "ok"})
	})

	// Auth endpoints
	ah := &AuthHandlers{Auth: d.AuthClient}
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", ah.Login)
		r.Post("/register", ah.Register)
		r.Post("/refresh", ah.Refresh)
		r.Get("/me", ah.Me)
	})

	// Room endpoints
	rh := &RoomHandlers{Room: d.RoomClient}
	r.Route("/rooms", func(rt chi.Router) {
		rt.Post("/", rh.CreateRoom)
		rt.Get("/", rh.ListRooms)

		rt.Route("/{id}", func(rr chi.Router) {
			rr.Get("/", rh.GetRoom)
			rr.Post("/join", rh.Join)
			rr.Post("/leave", rh.Leave)
			rr.Get("/participants", rh.Participants)
			rr.Get("/chat", rh.ChatHistory)
		})
	})

	return r
}
