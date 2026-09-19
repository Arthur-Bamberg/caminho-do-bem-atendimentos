package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVincularSegundoAssistidoAoNucleo(t *testing.T) {
	h := setup(t)
	a := postCadastro(t, h, map[string]any{
		"nome":       "Irmão 1",
		"unidade_id": "centro",
		"nucleo": map[string]any{
			"responsavel_legal": "Maria",
			"whatsapp":          "11988887777",
			"cesta":             true,
			"endereco":          map[string]any{"logradouro": "Rua A", "cidade": "São Paulo"},
		},
	})
	nucleoID := a["nucleo"].(map[string]any)["id"].(string)
	b := postCadastro(t, h, map[string]any{
		"nome":       "Irmão 2",
		"unidade_id": "norte",
		"nucleo":     map[string]any{"id": nucleoID},
	})
	if a["id"] == b["id"] {
		t.Fatal("mesmo assistido")
	}
	if b["nucleo"].(map[string]any)["id"] != nucleoID {
		t.Fatal("não vinculou")
	}
	if b["nucleo"].(map[string]any)["whatsapp"] != "11988887777" || b["nucleo"].(map[string]any)["cesta"] != true {
		t.Fatalf("não herdou %#v", b["nucleo"])
	}
	if a["unidade_id"] == b["unidade_id"] {
		t.Fatal("unidades deveriam diferir")
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/nucleos/"+nucleoID, nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("nucleo %d %s", rec.Code, rec.Body.String())
	}
	var n map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &n); err != nil {
		t.Fatal(err)
	}
	if n["assistidos"] != float64(2) {
		t.Fatalf("assistidos %v", n["assistidos"])
	}
	lista := listar(t, h)
	nucleos := map[string]bool{}
	for _, item := range lista {
		nucleos[item["nucleo_id"].(string)] = true
	}
	if len(lista) != 2 || len(nucleos) != 1 {
		t.Fatalf("famílias=%d assistidos=%d", len(nucleos), len(lista))
	}
}

func TestSemVinculoCriaNucleoDeUm(t *testing.T) {
	h := setup(t)
	a := postCadastro(t, h, map[string]any{"nome": "A"})
	b := postCadastro(t, h, map[string]any{"nome": "B"})
	if a["nucleo_id"] == b["nucleo_id"] {
		t.Fatal("fundiu núcleo")
	}
}

func TestHomonimosNaoFundemPorResponsavel(t *testing.T) {
	h := setup(t)
	a := postCadastro(t, h, map[string]any{"nucleo": map[string]any{"responsavel_legal": "Maria"}})
	b := postCadastro(t, h, map[string]any{"nucleo": map[string]any{"responsavel_legal": "Maria"}})
	if a["nucleo_id"] == b["nucleo_id"] {
		t.Fatal("fundiu por texto de responsável")
	}
}

func TestBuscaNucleoExplicita(t *testing.T) {
	h := setup(t)
	_ = postCadastro(t, h, map[string]any{"nucleo": map[string]any{"responsavel_legal": "Joana"}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/nucleos?q=joa", nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Itens []map[string]any `json:"itens"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Itens) != 1 || body.Itens[0]["responsavel_legal"] != "Joana" {
		t.Fatalf("%#v", body.Itens)
	}
}

func listar(t *testing.T, h http.Handler) []map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/assistidos", nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	var body struct {
		Itens []map[string]any `json:"itens"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body.Itens
}
