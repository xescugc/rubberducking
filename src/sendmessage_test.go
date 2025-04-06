package src_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xescugc/rubberducking/src"
)

var dataFile = path.Join(xdg.DataHome, src.AppName, "data.json")

func initializeData(t *testing.T, fs afero.Fs, surl string) {
	t.Helper()

	require.NoError(t, fs.MkdirAll(filepath.Dir(dataFile), 0700))

	data := src.Data{ManagerURL: surl}
	b, err := json.Marshal(data)
	require.NoError(t, err)

	err = afero.WriteFile(fs, dataFile, b, 0644)
	require.NoError(t, err)
}

func TestSendMessage(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			ctx = context.Background()
			fs  = afero.NewMemMapFs()
			msg = "Hi!"
		)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer ts.Close()

		initializeData(t, fs, ts.URL)

		err := src.SendMessage(ctx, fs, msg)
		require.NoError(t, err)
	})
	t.Run("Failed", func(t *testing.T) {
		t.Run("NoFile", func(t *testing.T) {
			var (
				ctx = context.Background()
				fs  = afero.NewMemMapFs()
				msg = "Hi!"
			)

			err := src.SendMessage(ctx, fs, msg)
			assert.EqualError(t, err, "there is no Duck running!")
		})
		t.Run("HTTPError", func(t *testing.T) {
			var (
				ctx = context.Background()
				fs  = afero.NewMemMapFs()
				msg = "Hi!"
			)

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				em := src.ErrorResponse{Error: "error"}
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(em)
			}))
			defer ts.Close()

			initializeData(t, fs, ts.URL)

			err := src.SendMessage(ctx, fs, msg)
			assert.EqualError(t, err, "error")
		})
	})
}
