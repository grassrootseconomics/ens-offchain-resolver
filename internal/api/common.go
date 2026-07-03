package api

import (
	"net/http"

	"github.com/kamikazechaser/common/httputil"
	"github.com/rs/cors"
	"github.com/uptrace/bunrouter"
)

func notFoundHandler(w http.ResponseWriter, _ bunrouter.Request) error {
	return httputil.JSON(w, http.StatusNotFound, ErrResponse{
		Ok:          false,
		Description: "Not found",
	})
}

func methodNotAllowedHandler(w http.ResponseWriter, _ bunrouter.Request) error {
	return httputil.JSON(w, http.StatusMethodNotAllowed, ErrResponse{
		Ok:          false,
		Description: "Method not allowed",
	})
}

func newCorsMiddleware(allowedOrigins []string) bunrouter.MiddlewareFunc {
	if len(allowedOrigins) == 0 {
		return func(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
			return next
		}
	}

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Accept", "Origin"},
		MaxAge:           86400,
	})

	return func(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
		return bunrouter.HTTPHandler(corsHandler.Handler(next))
	}
}
