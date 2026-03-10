package api

import (
	"io/fs"
	"net/http"
	"strings"

	"dst-ds-panel/internal/config"
	"dst-ds-panel/internal/docker"
	"dst-ds-panel/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Handler struct {
	docker  *docker.Manager
	store   *store.Store
	dataDir string
}

func NewHandler(dockerMgr *docker.Manager, store *store.Store, dataDir string) *Handler {
	return &Handler{
		docker:  dockerMgr,
		store:   store,
		dataDir: dataDir,
	}
}

func NewRouter(h *Handler, auth config.Auth, frontendFS fs.FS) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	authHandler := NewAuthHandler(auth)

	r.Route("/api", func(r chi.Router) {
		// Public: login endpoint
		r.Post("/login", authHandler.Login)

		// Protected: all other API routes
		r.Group(func(r chi.Router) {
			r.Use(authHandler.Middleware)

			r.Route("/clusters", func(r chi.Router) {
				r.Get("/", h.ListClusters)
				r.Post("/", h.CreateCluster)
				r.Post("/import", h.ImportCluster)

				r.Route("/{clusterID}", func(r chi.Router) {
					r.Get("/", h.GetCluster)
					r.Put("/config", h.UpdateClusterConfig)
					r.Delete("/", h.DeleteCluster)

					r.Post("/start", h.StartCluster)
					r.Post("/stop", h.StopCluster)
					r.Post("/restart", h.RestartCluster)
					r.Post("/clone", h.CloneCluster)

					r.Route("/shards/{shard}", func(r chi.Router) {
						r.Get("/logs", h.StreamLogs)
						r.Get("/stats", h.StreamStats)
						r.Post("/console", h.SendConsoleCommand)
					})

					r.Get("/mods", h.ListMods)
					r.Put("/mods", h.UpdateMods)

					r.Get("/backup", h.BackupCluster)
					r.Get("/players", h.GetPlayerActivity)

					r.Get("/files", h.ReadFile)
					r.Get("/files/list", h.ListFiles)
					r.Put("/files", h.WriteFile)
				})
			})

			r.Post("/image/build", h.BuildImage)
			r.Get("/image/status", h.ImageStatus)
			r.Post("/dst/update", h.UpdateDST)
		})
	})

	// Serve embedded frontend for all non-API routes (SPA fallback)
	if frontendFS != nil {
		fileServer := http.FileServer(http.FS(frontendFS))
		r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
			path := strings.TrimPrefix(req.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}
			if f, err := frontendFS.Open(path); err == nil {
				f.Close()
				fileServer.ServeHTTP(w, req)
				return
			}
			req.URL.Path = "/"
			fileServer.ServeHTTP(w, req)
		})
	}

	return r
}
