package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNucleoPersisteEVoltaNaFicha(t *testing.T) {
	h := setup(t)
	c := postCadastro(t, h, map[string]any{
		"nome": "Criança",
		"nucleo": map[string]any{
			"responsavel_legal": "Maria",
			"whatsapp":          "11999999999",
			"cesta":             true,
			"endereco": map[string]any{
				"logradouro": "Rua A",
				"numero":     "10",
				"bairro":     "Centro",
				"cidade":     "São Paulo",
				"uf":         "SP",
				"cep":        "01001000",
			},
		},
	})
	n := c["nucleo"].(map[string]any)
	if n["responsavel_legal"] != "Maria" || n["whatsapp"] != "11999999999" || n["cesta"] != true {
		t.Fatalf("núcleo %#v", n)
	}
	end := n["endereco"].(map[string]any)
	if end["logradouro"] != "Rua A" || end["cidade"] != "São Paulo" {
		t.Fatalf("endereço %#v", end)
	}
	lido := getCadastro(t, h, c["id"].(string))
	ln := lido["nucleo"].(map[string]any)
	if ln["responsavel_legal"] != "Maria" || ln["cesta"] != true {
		t.Fatalf("leitura %#v", ln)
	}
}

func TestDesmarcarCestaRefleteNaLeitura(t *testing.T) {
	h := setup(t)
	c := postCadastro(t, h, map[string]any{"nucleo": map[string]any{"cesta": true}})
	id := c["id"].(string)
	depois := putCadastro(t, h, id, map[string]any{"nucleo": map[string]any{"cesta": false}})
	if depois["nucleo"].(map[string]any)["cesta"] != false {
		t.Fatal("cesta deveria ser false")
	}
	lido := getCadastro(t, h, id)
	if lido["nucleo"].(map[string]any)["cesta"] != false {
		t.Fatal("leitura ainda com cesta")
	}
}

func TestResponsavelNaoCriaAssistido(t *testing.T) {
	h := setup(t)
	_ = postCadastro(t, h, map[string]any{
		"nome":   "João",
		"nucleo": map[string]any{"responsavel_legal": "Ana"},
	})
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
	if len(body.Itens) != 1 {
		t.Fatalf("assistidos %d", len(body.Itens))
	}
	if body.Itens[0]["nome"] != "João" {
		t.Fatalf("nome %v", body.Itens[0]["nome"])
	}
}

func TestNucleoCamposOpcionais(t *testing.T) {
	h := setup(t)
	c := postCadastro(t, h, map[string]any{})
	n := c["nucleo"].(map[string]any)
	if n["cesta"] != false || n["responsavel_legal"] != "" {
		t.Fatalf("%#v", n)
	}
}
