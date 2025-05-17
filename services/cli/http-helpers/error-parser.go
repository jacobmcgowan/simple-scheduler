package httpHelpers

import (
	"fmt"
	"io"
	"net/http"
)

func ParseError(resp *http.Response, msg string) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s with status %s", msg, resp.Status)
	}

	return fmt.Errorf("%s with status %s: %s", msg, resp.Status, string(body[:]))
}
