package api

import (
	"net/http"
	"strings"

	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
)

type CCIPParams struct {
	Sender string
	Data   string
}

func (a *API) ccipHandler(w http.ResponseWriter, req bunrouter.Request) error {
	r := CCIPParams{
		Sender: req.Param("sender"),
	}
	r.Data = strings.TrimSuffix(req.Param("data"), ".json")

	return httputil.JSON(w, http.StatusOK, OKResponse{
		Ok:          true,
		Description: "CCIP Data",
		Result: map[string]any{
			"sender": r.Sender,
			"data":   r.Data,
		},
	})
}
