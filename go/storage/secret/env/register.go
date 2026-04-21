package env

import (
	"fmt"

	"hop.top/kit/go/storage/secret"
)

func init() {
	secret.RegisterBackend("env", func(_ secret.Config) (secret.MutableStore, error) {
		return nil, fmt.Errorf("secret: env backend is read-only; use Store directly")
	})
}
