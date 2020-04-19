package providers

import (
	"net/http"
)

type ProviderSDUT struct {
	client *http.Client
}
