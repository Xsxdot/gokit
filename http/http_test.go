package http_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	gohttp "gokit/http"
)

// ---- helpers ----------------------------------------------------------------

func newEchoServer(t *testing.T, statusCode int, body string, extraHeaders map[string][]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, vs := range extraHeaders {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	}))
}

// ---- Fix 1: large-integer precision in query params -----------------------

func TestQueryParamsStruct_LargeInt(t *testing.T) {
	type Req struct {
		UserID int64  `json:"user_id"`
		Name   string `json:"name"`
	}

	var captured *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r
		w.WriteHeader(200)
	}))
	defer srv.Close()

	gohttp.Get(srv.URL).QueryParamsStruct(Req{UserID: 12345678901234, Name: "test"}).Do()

	got := captured.URL.Query().Get("user_id")
	if got != "12345678901234" {
		t.Errorf("large int precision lost: got %q, want %q", got, "12345678901234")
	}
}

// ---- Fix 2: Headers() preserves multi-values (e.g. Set-Cookie) -----------

func TestHeaders_MultiValue(t *testing.T) {
	srv := newEchoServer(t, 200, `{}`, map[string][]string{
		"Set-Cookie": {"session=abc; Path=/", "csrf=xyz; Path=/"},
	})
	defer srv.Close()

	resp := gohttp.Get(srv.URL).Do()

	headers := resp.Headers()
	cookies := headers["Set-Cookie"]
	if len(cookies) != 2 {
		t.Errorf("expected 2 Set-Cookie values, got %d: %v", len(cookies), cookies)
	}
}

// ---- Fix 2b: HeadersFlat() still works for single-value callers ----------

func TestHeadersFlat_SingleValue(t *testing.T) {
	srv := newEchoServer(t, 200, `{}`, nil)
	defer srv.Close()

	resp := gohttp.Get(srv.URL).Do()
	flat := resp.HeadersFlat()

	ct := flat["Content-Type"]
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("unexpected Content-Type: %q", ct)
	}
}

// ---- Fix 3: Options.Clone() prevents data race under concurrency ----------

func TestOptions_NoConcurrentDataRace(t *testing.T) {
	srv := newEchoServer(t, 200, `{}`, nil)
	defer srv.Close()

	client := gohttp.NewClient(nil)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			client.Get(srv.URL).
				Header("X-Worker", "yes").
				Do()
		}(i)
	}
	wg.Wait()
	// If there is a data race, -race will catch it and the test will fail.
}

// ---- Fix 4 (perf): Gson() caching — result is stable across calls --------

func TestGson_Caching(t *testing.T) {
	body := `{"code":0,"msg":"ok","data":{"id":1}}`
	srv := newEchoServer(t, 200, body, nil)
	defer srv.Close()

	resp := gohttp.Get(srv.URL).Do()

	r1 := resp.Gson().Get("data.id").Int()
	r2 := resp.Gson().Get("data.id").Int()
	if r1 != r2 {
		t.Errorf("Gson() not stable: first=%d second=%d", r1, r2)
	}
	if r1 != 1 {
		t.Errorf("unexpected data.id: %d", r1)
	}
}

// ---- Fix 5 (perf): Execute path — all HTTP methods work ------------------

func TestExecute_AllMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	client := gohttp.NewClient(nil)
	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			var resp *gohttp.Response
			switch m {
			case "GET":
				resp = client.Get(srv.URL).Do()
			case "POST":
				resp = client.Post(srv.URL).Do()
			case "PUT":
				resp = client.Put(srv.URL).Do()
			case "DELETE":
				resp = client.Delete(srv.URL).Do()
			case "PATCH":
				resp = client.Patch(srv.URL).Do()
			case "HEAD":
				resp = client.Head(srv.URL).Do()
			case "OPTIONS":
				resp = client.Options(srv.URL).Do()
			}
			if !resp.IsOK() {
				t.Errorf("%s failed: %v", m, resp.Err())
			}
		})
	}
}

// ---- Logic fix: assert error contains request method + URL ---------------

func TestAssertError_ContainsRequestURL(t *testing.T) {
	srv := newEchoServer(t, 404, `{"error":"not found"}`, nil)
	defer srv.Close()

	resp := gohttp.Get(srv.URL).Do().EnsureStatusCode(200)
	err := resp.Err()
	if err == nil {
		t.Fatal("expected assertion error, got nil")
	}
	if !strings.Contains(err.Error(), "GET") {
		t.Errorf("error message missing request method, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), srv.URL) {
		t.Errorf("error message missing request URL, got: %s", err.Error())
	}
}

// ---- Sanity: QueryParamsStruct matches JSON() on GET for simple structs --

func TestQueryParamsStruct_Equivalence(t *testing.T) {
	type Params struct {
		Page int    `json:"page"`
		Size int    `json:"size"`
		Tag  string `json:"tag"`
	}

	captured := make([]url.Values, 2)
	idx := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured[idx] = r.URL.Query()
		idx++
		w.WriteHeader(200)
	}))
	defer srv.Close()

	p := Params{Page: 1, Size: 20, Tag: "go"}

	gohttp.Get(srv.URL).JSON(p).Do()
	gohttp.Get(srv.URL).QueryParamsStruct(p).Do()

	// Both should produce the same query params
	for k, v := range captured[0] {
		if captured[1].Get(k) != v[0] {
			t.Errorf("key %s: JSON=%v QueryParamsStruct=%v", k, v, captured[1][k])
		}
	}
}
