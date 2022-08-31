package routes

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/RHEnVision/provisioning-backend/api"
	"github.com/RHEnVision/provisioning-backend/internal/middleware"
	s "github.com/RHEnVision/provisioning-backend/internal/services"
	"github.com/go-chi/chi/v5"
	redoc "github.com/go-openapi/runtime/middleware"
	"github.com/redhatinsights/platform-go-middlewares/identity"
)

func redocMiddleware(handler http.Handler) http.Handler {
	opt := redoc.RedocOpts{
		SpecURL: fmt.Sprintf("%s/openapi.json", PathPrefix()),
	}
	return redoc.Redoc(opt, handler)
}

func logETags() {
	logger := log.Logger
	for _, etag := range middleware.AllETags() {
		logger.Debug().Msgf("Calculated '%s' etag '%s' in %dms", etag.Name, etag.Value, etag.HashTime.Milliseconds())
	}
}

func SetupRoutes(r *chi.Mux) {
	r.Get("/ping", s.StatusService)
	r.Route("/docs", func(r chi.Router) {
		r.Use(redocMiddleware)
		r.Route("/openapi.json", func(r chi.Router) {
			r.Use(middleware.ETagMiddleware(api.ETagValue))
			r.Get("/", api.ServeOpenAPISpec)
		})
	})
	r.Mount(PathPrefix(), apiRouter())

	logETags()
}

func apiRouter() http.Handler {
	r := chi.NewRouter()

	r.Route("/openapi.json", func(r chi.Router) {
		r.Use(middleware.ETagMiddleware(api.ETagValue))
		r.Get("/", api.ServeOpenAPISpec)
	})
	r.Group(func(r chi.Router) {
		r.Use(identity.EnforceIdentity)
		r.Use(middleware.AccountMiddleware)

		r.Route("/ready", func(r chi.Router) {
			r.Get("/", s.ReadyService)
			r.Route("/{SRV}", func(r chi.Router) {
				r.Get("/", s.ReadyBackendService)
			})
		})

		r.Route("/sources", func(r chi.Router) {
			r.Get("/", s.ListSources)
			r.Route("/{ID}", func(r chi.Router) {
				r.Get("/instance_types", s.ListInstanceTypes)
			})
		})

		r.Route("/pubkeys", func(r chi.Router) {
			r.Post("/", s.CreatePubkey)
			r.Get("/", s.ListPubkeys)
			r.Route("/{ID}", func(r chi.Router) {
				r.Get("/", s.GetPubkey)
				r.Delete("/", s.DeletePubkey)
			})
		})

		r.Route("/reservations", func(r chi.Router) {
			r.Get("/", s.ListReservations)
			r.Route("/{type}", func(r chi.Router) {
				r.Post("/", s.CreateReservation)
			})
		})
	})

	return r
}
