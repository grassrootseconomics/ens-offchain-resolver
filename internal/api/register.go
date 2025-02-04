package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
	"golang.org/x/exp/rand"
)

const (
	domainSuffix = ".sarafu.eth"
)

var (
	validSubdomain = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
)

func (a *API) registerHandler(w http.ResponseWriter, req bunrouter.Request) error {
	var registerReq RegisterRequest

	if err := a.validator.BindJSONAndValidate(w, req.Request, &registerReq); err != nil {
		a.logg.Error("validation failed", "error", err)
		return httputil.JSON(w, http.StatusBadRequest, ErrResponse{
			Ok:          false,
			Description: "Validation failed",
		})
	}

	subdomain, err := extractSubdomain(registerReq.Hint)
	if err != nil {
		return httputil.JSON(w, http.StatusBadRequest, ErrResponse{
			Ok:          false,
			Description: err.Error(),
		})
	}

	_, err = a.store.LookupName(req.Context(), registerReq.Hint)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if err := a.store.RegisterName(req.Context(), registerReq.Hint, registerReq.Address); err != nil {
				a.logg.Error("register failed", "error", err)
				return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
					Ok:          false,
					Description: "Internal server error",
				})
			}

			return httputil.JSON(w, http.StatusOK, OKResponse{
				Ok:          true,
				Description: "Name registered",
				Result: map[string]any{
					"address":    registerReq.Address,
					"name":       registerReq.Hint,
					"autoChoose": false,
				},
			})
		} else {
			a.logg.Error("lookup failed", "error", err)
			return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
				Ok:          false,
				Description: "Internal server error",
			})
		}
	}

	return a.autoChoose(req.Context(), subdomain, registerReq.Address, w)
}

func (a *API) autoChoose(ctx context.Context, subdomain string, address string, w http.ResponseWriter) error {
	// Max of 90 iterations to find the first available alias + suffix
	for i := 0; i < 90; i++ {
		a.logg.Debug("autochoose iteration", "iteration", i, "subdomain", subdomain)
		num := rand.Intn(90) + 10
		randName := fmt.Sprintf("%s%d", subdomain, num)
		_, err := a.store.LookupName(ctx, randName)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				if err := a.store.RegisterName(ctx, randName+domainSuffix, address); err != nil {
					var pgErr *pgconn.PgError
					if errors.As(err, &pgErr) && pgErr.Code == "23505" {
						continue
					}

					a.logg.Error("register failed", "error", err)
					return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
						Ok:          false,
						Description: "Internal server error",
					})
				}

				return httputil.JSON(w, http.StatusOK, OKResponse{
					Ok:          true,
					Description: "Name registered",
					Result: map[string]any{
						"address":    address,
						"name":       randName + domainSuffix,
						"autoChoose": true,
					},
				})
			}
		}
	}

	return httputil.JSON(w, http.StatusServiceUnavailable, ErrResponse{
		Ok:          false,
		Description: "Autochoose error, try a different hint",
	})
}

func extractSubdomain(hint string) (string, error) {
	parts := strings.Split(hint, ".")
	if len(parts) > 3 {
		return "", fmt.Errorf("invalid ENS name")
	}
	if !isValidSubdomain(parts[0]) {
		return "", fmt.Errorf("invalid subdomain format: only letters and numbers are allowed")
	}
	return parts[0], nil
}

func isValidSubdomain(subdomain string) bool {
	return validSubdomain.MatchString(subdomain)
}
