package csapi_tests

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/matrix-org/complement/internal/b"
	"github.com/matrix-org/complement/internal/client"
	"github.com/matrix-org/complement/internal/data"
	"github.com/matrix-org/complement/runtime"
)

func TestAsyncUpload(t *testing.T) {
	runtime.SkipIf(t, runtime.Dendrite) // Dendrite doesn't support async uploads

	deployment := Deploy(t, b.BlueprintAlice)
	defer deployment.Destroy(t)

	alice := deployment.Client(t, "hs1", "@alice:hs1")

	var mxcURI, mediaID string
	t.Run("Create media", func(t *testing.T) {
		mxcURI = alice.CreateMedia(t)
		parts := strings.Split(mxcURI, "/")
		mediaID = parts[len(parts)-1]
	})

	origin, mediaID := client.SplitMxc(mxcURI)

	t.Run("Not yet uploaded", func(t *testing.T) {
		// Check that the media is not yet uploaded
		res := alice.DoFunc(t, "GET", []string{"_matrix", "media", "v3", "download", origin, mediaID})
		if res.StatusCode != http.StatusGatewayTimeout {
			t.Fatalf("Expected 504 response code, got %d", res.StatusCode)
		}
	})

	wantContentType := "image/png"

	t.Run("Upload media", func(t *testing.T) {
		alice.UploadMediaAsync(t, origin, mediaID, data.MatrixPng, "test.png", wantContentType)
	})

	t.Run("Cannot upload to a media ID that has already been uploaded to", func(t *testing.T) {
		alice.UploadMediaAsync(t, origin, mediaID, data.MatrixPng, "test.png", wantContentType)
	})

	t.Run("Download media", func(t *testing.T) {
		content, contentType := alice.DownloadContent(t, mxcURI)
		if !bytes.Equal(data.MatrixPng, content) {
			t.Fatalf("uploaded and downloaded content doesn't match: want %v\ngot\n%v", data.MatrixPng, content)
		}
		if contentType != wantContentType {
			t.Fatalf("expected contentType to be %s, got %s", wantContentType, contentType)
		}
	})
}