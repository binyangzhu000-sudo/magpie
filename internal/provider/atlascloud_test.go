package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAtlasCloudPreset(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	p, err := FromPreset("atlascloud")
	if err != nil {
		t.Fatal(err)
	}
	if pr := Preset("atlascloud"); pr.Kind != KindRelay || pr.NoKey || p.Icon != "atlascloud" || p.KeysURL != "https://console.atlascloud.ai/api-keys" {
		t.Fatalf("Atlas Cloud must be a keyed relay with its icon and key page: %+v", pr)
	}
	if p.Chat != "https://api.atlascloud.ai/v1" || p.Responses != p.Chat || p.Anthropic != "https://api.atlascloud.ai" {
		t.Fatalf("Atlas Cloud endpoints: %q %q %q", p.Chat, p.Responses, p.Anthropic)
	}
	// An existing OpenAI-compatible entry is recognized on import and gains
	// the preset's other APIs, with the user's key kept.
	p, why := imported("Atlas Cloud", "atlas-key", endpoints{chat: p.Chat}, nil)
	if why != "" || p.Preset != "atlascloud" || p.Key != "atlas-key" || p.Responses == "" || p.Anthropic == "" {
		t.Fatalf("Atlas Cloud import: %+v, %s", p, why)
	}
	// Its /v1/models wraps the list as {code, msg, data}, the ids vendor/model.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" || r.Header.Get("Authorization") != "Bearer atlas-key" {
			t.Errorf("model request: %s, auth %q", r.URL.Path, r.Header.Get("Authorization"))
			http.Error(w, "bad model request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":200,"msg":"succeed","data":[
			{"id":"deepseek-ai/deepseek-v4-flash","object":"model","context_length":1048576},
			{"id":"moonshotai/kimi-k3","object":"model","context_length":262144}
		]}`))
	}))
	defer srv.Close()
	p.Chat, p.Responses, p.Anthropic = srv.URL+"/v1", srv.URL+"/v1", srv.URL
	ms, err := p.Fetch(context.Background())
	if err != nil || len(ms) != 2 || ms[0].ID != "deepseek-ai/deepseek-v4-flash" {
		t.Fatalf("Atlas Cloud models: %+v, %v", ms, err)
	}
}
