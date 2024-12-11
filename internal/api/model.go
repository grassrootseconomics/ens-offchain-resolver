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

	RegisterRequest struct {
		Address string `json:"address" validate:"required,eth_addr_checksum"`
		Name    string `json:"name" validate:"required,fqdn"`
	}
)
