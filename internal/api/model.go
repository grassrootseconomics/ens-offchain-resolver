package api

type (
	OKResponse struct {
		Ok          bool           `json:"ok"`
		Description string         `json:"description"`
		Result      map[string]any `json:"result"`
	}

	ErrResponse struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}

	CCIPOKResponse struct {
		Data string `json:"data"`
	}

	CCIPErrResponse struct {
		Message string `json:"message"`
	}

	RegisterRequest struct {
		Address string `json:"address" validate:"required,eth_addr_checksum"`
		Hint    string `json:"hint" validate:"required,fqdn"`
	}

	UpdateRequest struct {
		Name    string `json:"name" validate:"required,fqdn"`
		Address string `json:"address" validate:"required,eth_addr_checksum"`
	}
)
