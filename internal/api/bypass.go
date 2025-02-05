package api

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
)

func (a *API) resolveHandler(w http.ResponseWriter, req bunrouter.Request) error {
	q := req.URL.Query()

	if len(q["name"]) == 1 {
		address, err := a.store.LookupName(req.Context(), q["name"][0])
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return httputil.JSON(w, http.StatusNotFound, ErrResponse{
					Ok:          false,
					Description: "Name not found",
				})
			}

			return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
				Ok:          false,
				Description: "Internal server error",
			})
		}

		return httputil.JSON(w, http.StatusOK, OKResponse{
			Ok:          true,
			Description: "Name resolved",
			Result: map[string]any{
				"address": address,
			},
		})
	}

	if len(q["address"]) == 1 {
		name, err := a.store.ReverseLookup(req.Context(), q["address"][0])
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return httputil.JSON(w, http.StatusNotFound, ErrResponse{
					Ok:          false,
					Description: "Address not found",
				})
			}

			return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
				Ok:          false,
				Description: "Internal server error",
			})
		}

		return httputil.JSON(w, http.StatusOK, OKResponse{
			Ok:          true,
			Description: "Name resolved",
			Result: map[string]any{
				"name": name,
			},
		})
	}

	return httputil.JSON(w, http.StatusBadRequest, ErrResponse{
		Ok:          false,
		Description: "Invalid query param",
	})
}
