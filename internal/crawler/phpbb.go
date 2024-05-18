package crawler

import (
	"errors"
	"fmt"
	"net/http"
)

func FetchUserRoles(userID string) ([]string, error) {
	url := fmt.Sprintf("")
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.New("could not fetch user roles")
	}
	defer resp.Body.Close()

	return nil, nil
}
