package api

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
)

func (a *API) resolveNameHandler(w http.ResponseWriter, req bunrouter.Request) error {
	// TODO: Validation?
	name := req.Param("name")

	address, err := a.store.LookupName(req.Context(), name)
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

func (a *API) reverseAddressHandler(w http.ResponseWriter, req bunrouter.Request) error {
	address := req.Param("address")

	name, err := a.store.ReverseLookup(req.Context(), address)
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
