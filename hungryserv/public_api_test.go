package hungryserv

import (
	"net/url"
	"testing"

	"github.com/matrix-org/complement/internal/b"
	"github.com/matrix-org/complement/internal/client"
	"github.com/matrix-org/complement/internal/match"
	"github.com/matrix-org/complement/internal/must"
)

// Test all public APIs do what they say
func TestPublicAPI(t *testing.T) {
	deployment := Deploy(t, b.BlueprintCleanHS)
	defer deployment.Destroy(t)

	anon := deployment.Client(t, "hs1", "")

	t.Run("Registration API must not be open", func(t *testing.T) {
		// Guest
		resp := anon.DoFunc(t, "POST", []string{"_matrix", "client", "v3", "register"}, client.WithQueries(url.Values{"kind": []string{"guest"}}), client.WithJSONBody(t, map[string]interface{}{}))
		must.MatchResponse(t, resp, match.HTTPResponse{
			StatusCode: 401,
		})

		// User
		resp = anon.DoFunc(t, "POST", []string{"_matrix", "client", "v3", "register"}, client.WithQueries(url.Values{"kind": []string{"user"}}), client.WithJSONBody(t, map[string]interface{}{
			"username": "foo",
			"password": "bar",
		}))
		must.MatchResponse(t, resp, match.HTTPResponse{
			StatusCode: 401,
		})
	})

	t.Run("Login must send back client error for invalid payload", func(t *testing.T) {
		resp := anon.DoFunc(t, "POST", []string{"_matrix", "client", "v3", "login"}, client.WithJSONBody(t, map[string]interface{}{}))
		must.MatchResponse(t, resp, match.HTTPResponse{
			StatusCode: 400,
		})
	})

	t.Run("Login must not work with wrong credentials", func(t *testing.T) {
		resp := anon.DoFunc(t, "POST", []string{"_matrix", "client", "v3", "login"}, client.WithJSONBody(t, map[string]interface{}{
			"identifier": map[string]interface{}{
				"type": "m.id.user",
				"user": "@foo:hs1",
			},
			"password": "bar",
			"type":     "m.login.password",
		}))
		must.MatchResponse(t, resp, match.HTTPResponse{
			StatusCode: 403,
		})
	})

	t.Run("Appservice login must not work with wrong credentials", func(t *testing.T) {
		resp := anon.DoFunc(t, "POST", []string{"_matrix", "client", "v3", "login"}, client.WithJSONBody(t, map[string]interface{}{
			"identifier": map[string]interface{}{
				"type": "m.id.user",
				"user": "@foo:hs1",
			},
			"type": "m.login.application_service",
		}))
		must.MatchResponse(t, resp, match.HTTPResponse{
			StatusCode: 401,
		})
	})

	t.Run("Proxied media endpoints must not require authentication", func(t *testing.T) {
		// We don't have another HS running so GW error is expected for these
		resp := anon.DoFunc(t, "GET", []string{"_matrix", "media", "v3", "download", "hs1", "non-existing-media"})
		must.MatchResponse(t, resp, match.HTTPResponse{
			StatusCode: 502,
		})

		resp = anon.DoFunc(t, "GET", []string{"_matrix", "media", "v3", "download", "hs1", "non-existing-media", ""})
		must.MatchResponse(t, resp, match.HTTPResponse{
			StatusCode: 502,
		})

		resp = anon.DoFunc(t, "GET", []string{"_matrix", "media", "v3", "download", "hs1", "non-existing-media", "non-existing-filename"})
		must.MatchResponse(t, resp, match.HTTPResponse{
			StatusCode: 502,
		})

		resp = anon.DoFunc(t, "GET", []string{"_matrix", "media", "v3", "thumbnail", "hs1", "non-existing-media"})
		must.MatchResponse(t, resp, match.HTTPResponse{
			StatusCode: 502,
		})
	})

	t.Run("Asmux AS endpoints must not work without bearer", func(t *testing.T) {
		resp := anon.DoFunc(t, "GET", []string{"_matrix", "asmux", "appservice", "alice", "bridge"})
		must.MatchResponse(t, resp, match.HTTPResponse{
			StatusCode: 403,
		})

		resp = anon.DoFunc(t, "PUT", []string{"_matrix", "asmux", "appservice", "alice", "bridge"}, client.WithJSONBody(t, map[string]interface{}{}))
		must.MatchResponse(t, resp, match.HTTPResponse{
			StatusCode: 403,
		})

		resp = anon.DoFunc(t, "DELETE", []string{"_matrix", "asmux", "appservice", "alice", "bridge"})
		must.MatchResponse(t, resp, match.HTTPResponse{
			StatusCode: 403,
		})
	})
}
