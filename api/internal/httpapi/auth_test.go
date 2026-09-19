package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Arthur-Bamberg/nacao-assistidos/api/internal/store"
)

const emailSeed = "operador@nacao.local"
const senhaSeed = "nacao-dev"

func setup(t *testing.T) http.Handler {
	t.Helper()
	mem := store.NewMemory()
	if err := mem.SeedOperador(emailSeed, senhaSeed); err != nil {
		t.Fatal(err)
	}
	return New(mem).Handler()
}

func TestLoginValidoGravaCookieHttpOnly(t *testing.T) {
	h := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login", jsonBody(map[string]string{
		"email": emailSeed,
		"senha": senhaSeed,
	}))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d corpo %s", rec.Code, rec.Body.String())
	}
	c := cookieSessaoDe(t, rec)
	if !c.HttpOnly {
		t.Fatal("cookie precisa ser httpOnly")
	}
	if c.Value == "" {
		t.Fatal("cookie sem token")
	}
}

func TestLoginInvalido(t *testing.T) {
	h := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login", jsonBody(map[string]string{
		"email": emailSeed,
		"senha": "errada",
	}))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestRotaAutenticadaVsAnonima(t *testing.T) {
	h := setup(t)

	anon := httptest.NewRecorder()
	h.ServeHTTP(anon, httptest.NewRequest(http.MethodGet, "/api/me", nil))
	if anon.Code != http.StatusUnauthorized {
		t.Fatalf("anônimo me: %d", anon.Code)
	}
	for _, path := range []string{"/api/assistidos", "/api/indicadores"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("anônimo %s: %d", path, rec.Code)
		}
	}

	cookie := loginCookie(t, h)
	aut := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.AddCookie(cookie)
	h.ServeHTTP(aut, req)
	if aut.Code != http.StatusOK {
		t.Fatalf("autenticado me: %d %s", aut.Code, aut.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(aut.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["email"] != emailSeed {
		t.Fatalf("email %q", body["email"])
	}
}

func TestLogoutEncerraSessao(t *testing.T) {
	h := setup(t)
	cookie := loginCookie(t, h)

	out := httptest.NewRecorder()
	reqOut := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	reqOut.AddCookie(cookie)
	h.ServeHTTP(out, reqOut)
	if out.Code != http.StatusNoContent {
		t.Fatalf("logout %d", out.Code)
	}

	depois := httptest.NewRecorder()
	reqMe := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	reqMe.AddCookie(cookie)
	h.ServeHTTP(depois, reqMe)
	if depois.Code != http.StatusUnauthorized {
		t.Fatalf("após logout %d", depois.Code)
	}
}

func TestSaudeNaoExigeSessao(t *testing.T) {
	h := setup(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/saude", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("saude %d", rec.Code)
	}
}

func loginCookie(t *testing.T, h http.Handler) *http.Cookie {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login", jsonBody(map[string]string{
		"email": emailSeed,
		"senha": senhaSeed,
	}))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login %d %s", rec.Code, rec.Body.String())
	}
	return cookieSessaoDe(t, rec)
}

func cookieSessaoDe(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	res := rec.Result()
	defer res.Body.Close()
	for _, c := range res.Cookies() {
		if c.Name == cookieSessao {
			return c
		}
	}
	t.Fatal("cookie de sessão ausente")
	return nil
}

func jsonBody(v any) io.Reader {
	b, _ := json.Marshal(v)
	return bytes.NewReader(b)
}

func TestLoginNormalizaEmail(t *testing.T) {
	h := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login", jsonBody(map[string]string{
		"email": strings.ToUpper(emailSeed),
		"senha": senhaSeed,
	}))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}
