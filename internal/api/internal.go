package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
)

type PublicAddressParam struct {
	Address string `validate:"required,eth_addr_checksum"`
}

func (a *API) resolveHandler(w http.ResponseWriter, req bunrouter.Request) error {
	name := strings.ToLower(req.Param("name"))
	if name == "" {
		return httputil.JSON(w, http.StatusBadRequest, ErrResponse{
			Ok:          false,
			Description: "Name parameter is required",
		})
	}

	address, err := a.store.LookupName(req.Context(), name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			resolvedAddress, err := a.ensProvider.ResolveName(name)
			if err != nil {
				return httputil.JSON(w, http.StatusNotFound, ErrResponse{
					Ok:          false,
					Description: "Name not found",
				})
			}

			return httputil.JSON(w, http.StatusOK, OKResponse{
				Ok:          true,
				Description: "Address resolved",
				Result: map[string]any{
					"address": resolvedAddress.Hex(),
				},
			})
		}

		return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
			Ok:          false,
			Description: "Internal server error",
		})
	}

	return httputil.JSON(w, http.StatusOK, OKResponse{
		Ok:          true,
		Description: "Address resolved",
		Result: map[string]any{
			"address": address,
		},
	})
}

func (a *API) reverseResolveHandler(w http.ResponseWriter, req bunrouter.Request) error {
	r := PublicAddressParam{
		Address: req.Param("address"),
	}

	if err := a.validator.Validate(r); err != nil {
		return httputil.JSON(w, http.StatusBadRequest, ErrResponse{
			Ok:          false,
			Description: "Address validation failed",
		})
	}

	name, err := a.store.ReverseLookup(req.Context(), r.Address)
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
		Description: "Name reverse resolved",
		Result: map[string]any{
			"name": name,
		},
	})
}
