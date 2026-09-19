package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConsultaBuscaFiltroPaginacao(t *testing.T) {
	h := setup(t)
	_ = postCadastro(t, h, map[string]any{"nome": "Ana", "unidade_id": "centro", "cpf": "52998224725", "nucleo": map[string]any{"responsavel_legal": "Maria"}})
	_ = postCadastro(t, h, map[string]any{"nome": "Bruno", "unidade_id": "norte", "nucleo": map[string]any{"responsavel_legal": "João"}})
	_ = postCadastro(t, h, map[string]any{"nome": "Carla", "unidade_id": "centro"})

	ana := consultar(t, h, "q=ana")
	if ana.Total != 1 || ana.Itens[0]["nome"] != "Ana" {
		t.Fatalf("nome %#v", ana)
	}
	cpf := consultar(t, h, "q=52998224725")
	if cpf.Total != 1 {
		t.Fatalf("cpf %d", cpf.Total)
	}
	resp := consultar(t, h, "q=maria")
	if resp.Total != 1 {
		t.Fatalf("responsável %d", resp.Total)
	}
	centro := consultar(t, h, "unidade_id=centro")
	if centro.Total != 2 {
		t.Fatalf("centro %d", centro.Total)
	}
	todas := consultar(t, h, "")
	if todas.Total != 3 {
		t.Fatalf("todas %d", todas.Total)
	}
	p1 := consultar(t, h, "por_pagina=1&pagina=1")
	p2 := consultar(t, h, "por_pagina=1&pagina=2")
	if p1.Total != 3 || len(p1.Itens) != 1 || len(p2.Itens) != 1 || p1.Itens[0]["id"] == p2.Itens[0]["id"] {
		t.Fatalf("paginação %#v %#v", p1, p2)
	}
	vazio := consultar(t, h, "q=xyz-nao-existe")
	if vazio.Total != 0 || len(vazio.Itens) != 0 {
		t.Fatalf("vazio %#v", vazio)
	}
}

func TestReabrirSalvaNovoAtendimento(t *testing.T) {
	h := setup(t)
	c := postCadastro(t, h, map[string]any{"nome": "Ana", "atendimento": map[string]any{"relato": "setembro"}})
	id := c["id"].(string)
	_ = putCadastro(t, h, id, map[string]any{"nome": "Ana Silva", "atendimento": map[string]any{"relato": "novembro"}})
	lido := getCadastro(t, h, id)
	if lido["nome"] != "Ana Silva" {
		t.Fatalf("nome vigente %v", lido["nome"])
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/assistidos/"+id+"/atendimentos", nil)
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
	if len(body.Itens) != 2 {
		t.Fatalf("atendimentos %d", len(body.Itens))
	}
	relatos := body.Itens[0]["relato"].(string) + body.Itens[1]["relato"].(string)
	if !strings.Contains(relatos, "setembro") || !strings.Contains(relatos, "novembro") {
		t.Fatalf("relatos %q", relatos)
	}
}

type consultaResp struct {
	Itens     []map[string]any `json:"itens"`
	Total     int              `json:"total"`
	Pagina    int              `json:"pagina"`
	PorPagina int              `json:"por_pagina"`
}

func consultar(t *testing.T, h http.Handler, qs string) consultaResp {
	t.Helper()
	path := "/api/assistidos"
	if qs != "" {
		path += "?" + qs
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	var body consultaResp
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body
}
