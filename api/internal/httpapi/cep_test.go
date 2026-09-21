package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Arthur-Bamberg/nacao-assistidos/api/internal/store"
)

func setupCEP(t *testing.T, via http.HandlerFunc) http.Handler {
	t.Helper()
	fake := httptest.NewServer(via)
	t.Cleanup(fake.Close)
	mem := store.NewMemory()
	if err := mem.SeedOperador(emailSeed, senhaSeed); err != nil {
		t.Fatal(err)
	}
	s := New(mem)
	s.viacepBase = fake.URL
	s.httpClient = fake.Client()
	return s.Handler()
}

func TestCEPAnonimo401(t *testing.T) {
	h := setup(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/cep/01001000", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestCEPInvalido400(t *testing.T) {
	h := setupCEP(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("não deveria chamar ViaCEP")
		w.WriteHeader(http.StatusOK)
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/cep/123", nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
}

func TestCEPConsultaViaCEP(t *testing.T) {
	h := setupCEP(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ws/01001000/json/" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"cep":        "01001-000",
			"logradouro": "Praça da Sé",
			"complemento": "lado ímpar",
			"bairro":     "Sé",
			"localidade": "São Paulo",
			"uf":         "SP",
		})
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/cep/01001-000", nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["logradouro"] != "Praça da Sé" || body["bairro"] != "Sé" || body["cidade"] != "São Paulo" || body["uf"] != "SP" {
		t.Fatalf("%#v", body)
	}
	if body["cep"] != "01001000" {
		t.Fatalf("cep %q", body["cep"])
	}
	if _, ok := body["complemento"]; ok {
		t.Fatalf("não deve devolver complemento do ViaCEP: %#v", body)
	}
}

func TestCEPNaoEncontrado404(t *testing.T) {
	h := setupCEP(t, func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"erro": true})
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/cep/00000000", nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
}

func TestCEPViaCEPIndisponivel502(t *testing.T) {
	h := setupCEP(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/cep/01001000", nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
}
