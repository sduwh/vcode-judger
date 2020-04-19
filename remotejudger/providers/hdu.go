package providers

import (
	"net/http"
)

type ProviderHDU struct {
	client *http.Client
}
