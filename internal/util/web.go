package util

import (
	"encoding/base64"
	"fmt"
)

func BasicAuth(username string, password string) string {
	credentials := fmt.Sprintf("%s:%s", username, password)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encodedCredentials)
}
