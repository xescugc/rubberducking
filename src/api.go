package src

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func postToGame(url string, body io.Reader) error {
	resp, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(resp)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		var eb ErrorResponse
		err := json.NewDecoder(response.Body).Decode(&eb)
		if err != nil {
			return err
		}
		return errors.New(eb.Error)
	}

	return nil
}
