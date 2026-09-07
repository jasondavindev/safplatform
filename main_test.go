package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func doRequest(t *testing.T, req *http.Request) (*httptest.ResponseRecorder, serverResponse) {
	t.Helper()

	rec := httptest.NewRecorder()
	newRouter().ServeHTTP(rec, req)

	var got serverResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v (body=%s)", err, rec.Body.String())
	}
	return rec, got
}

func TestEchoJSONBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/things?a=1&a=2&b=x", strings.NewReader(`{"name":"turbi","n":3}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Custom", "abc")

	rec, got := doRequest(t, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	if got.Path != "/api/v1/things" {
		t.Errorf("Path = %q, want /api/v1/things", got.Path)
	}
	if want := []string{"1", "2"}; !equal(got.Params["a"], want) {
		t.Errorf("Params[a] = %v, want %v", got.Params["a"], want)
	}
	if want := []string{"x"}; !equal(got.Params["b"], want) {
		t.Errorf("Params[b] = %v, want %v", got.Params["b"], want)
	}
	if want := []string{"abc"}; !equal(got.Headers["X-Custom"], want) {
		t.Errorf("Headers[X-Custom] = %v, want %v", got.Headers["X-Custom"], want)
	}

	body, ok := got.Body.(map[string]any)
	if !ok {
		t.Fatalf("Body is %T, want JSON object", got.Body)
	}
	if body["name"] != "turbi" {
		t.Errorf("Body[name] = %v, want turbi", body["name"])
	}
	if body["n"] != float64(3) {
		t.Errorf("Body[n] = %v, want 3", body["n"])
	}
}

func TestEchoPlainTextBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/plain", strings.NewReader("not json at all"))

	_, got := doRequest(t, req)

	if got.Body != "not json at all" {
		t.Errorf("Body = %#v, want raw string", got.Body)
	}
}

func TestEchoEmptyBodyAndNoParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec, got := doRequest(t, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if got.Body != "" {
		t.Errorf("Body = %#v, want empty string", got.Body)
	}
	if got.Path != "/" {
		t.Errorf("Path = %q, want /", got.Path)
	}
	if len(got.Params) != 0 {
		t.Errorf("Params = %v, want empty", got.Params)
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	for _, key := range []string{"Headers", "Params", "Body", "Path"} {
		if _, ok := envelope[key]; !ok {
			t.Errorf("response missing key %q", key)
		}
	}
	if string(envelope["Params"]) != "{}" {
		t.Errorf("Params serialized as %s, want {}", envelope["Params"])
	}
}

func TestEchoArbitraryPathsAndMethods(t *testing.T) {
	paths := []string{"/", "/a", "/a/b/c", "/deep/nested/route/with/many/segments"}
	methods := []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPatch, http.MethodOptions, "WEIRDVERB"}

	for _, path := range paths {
		for _, method := range methods {
			req := httptest.NewRequest(method, path, nil)
			rec, got := doRequest(t, req)

			if rec.Code != http.StatusOK {
				t.Errorf("%s %s: status = %d, want 200", method, path, rec.Code)
			}
			if got.Path != path {
				t.Errorf("%s %s: Path = %q", method, path, got.Path)
			}
		}
	}
}

func TestDecodeBody(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", `""`},
		{"object", `{"a":1}`, `{"a":1}`},
		{"array", `[1,2]`, `[1,2]`},
		{"number", `42`, `42`},
		{"quoted string", `"hi"`, `"hi"`},
		{"invalid json", `{broken`, `"{broken"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := json.Marshal(decodeBody([]byte(tc.in)))
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(out) != tc.want {
				t.Errorf("decodeBody(%q) marshalled to %s, want %s", tc.in, out, tc.want)
			}
		})
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
