package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestCatalogoUnidadesEOficinas(t *testing.T) {
	h := setup(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/catalogos", nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var body struct {
		Unidades []map[string]string `json:"unidades"`
		Oficinas []map[string]string `json:"oficinas"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Unidades) != 3 || len(body.Oficinas) != 6 {
		t.Fatalf("catálogo %#v", body)
	}
}

func TestGravaLeUnidadeOficinasEAluno(t *testing.T) {
	h := setup(t)
	c := postCadastro(t, h, map[string]any{
		"unidade_id": "centro",
		"oficinas":   []string{"jiu-jitsu", "nacao-esporte"},
	})
	if c["unidade_id"] != "centro" {
		t.Fatalf("unidade %v", c["unidade_id"])
	}
	if c["aluno"] != true {
		t.Fatal("deveria ser Aluno")
	}
	ids := oficinasDe(t, c)
	if !reflect.DeepEqual(ids, []string{"jiu-jitsu", "nacao-esporte"}) {
		t.Fatalf("oficinas %v", ids)
	}
	lido := getCadastro(t, h, c["id"].(string))
	if lido["unidade_id"] != "centro" || lido["aluno"] != true {
		t.Fatalf("leitura %#v", lido)
	}
}

func TestSemOficinaNaoEAluno(t *testing.T) {
	h := setup(t)
	c := postCadastro(t, h, map[string]any{"unidade_id": "norte"})
	if c["aluno"] != false {
		t.Fatal("sem oficina não é Aluno")
	}
}

func TestSalvarDeNovoSubstituiOficinas(t *testing.T) {
	h := setup(t)
	c := postCadastro(t, h, map[string]any{"oficinas": []string{"jiu-jitsu", "nacao-cultura"}})
	id := c["id"].(string)
	depois := putCadastro(t, h, id, map[string]any{"oficinas": []string{"acessuas-trabalho"}})
	ids := oficinasDe(t, depois)
	if !reflect.DeepEqual(ids, []string{"acessuas-trabalho"}) {
		t.Fatalf("conjunto %v", ids)
	}
	if depois["aluno"] != true {
		t.Fatal("ainda Aluno")
	}
	vazio := putCadastro(t, h, id, map[string]any{"oficinas": []string{}})
	if vazio["aluno"] != false {
		t.Fatal("sem oficina deixa de ser Aluno")
	}
}

func TestUnidadesDistintasEmAssistidosSeparados(t *testing.T) {
	h := setup(t)
	a := postCadastro(t, h, map[string]any{"unidade_id": "centro"})
	b := postCadastro(t, h, map[string]any{"unidade_id": "norte"})
	if a["unidade_id"] == b["unidade_id"] {
		t.Fatal("unidades iguais")
	}
}

func oficinasDe(t *testing.T, c map[string]any) []string {
	t.Helper()
	raw, _ := c["oficinas"].([]any)
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		out = append(out, v.(string))
	}
	return out
}

func getCadastro(t *testing.T, h http.Handler, id string) map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/assistidos/"+id, nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get %d %s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func putCadastro(t *testing.T, h http.Handler, id string, body map[string]any) map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/assistidos/"+id, jsonBody(body))
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("put %d %s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func postUnidade(t *testing.T, h http.Handler, name string) map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/unidades", jsonBody(map[string]any{"nome": name}))
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("postUnidade status %d %s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func putUnidade(t *testing.T, h http.Handler, id, name string) map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/unidades/"+id, jsonBody(map[string]any{"nome": name}))
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("putUnidade status %d %s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func putUnidadeAtivo(t *testing.T, h http.Handler, id string, ativo bool) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/unidades/"+id+"/ativo", jsonBody(map[string]any{"ativo": ativo}))
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("putUnidadeAtivo status %d %s", rec.Code, rec.Body.String())
	}
}

func postOficina(t *testing.T, h http.Handler, name string) map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/oficinas", jsonBody(map[string]any{"nome": name}))
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("postOficina status %d %s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func putOficina(t *testing.T, h http.Handler, id, name string) map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/oficinas/"+id, jsonBody(map[string]any{"nome": name}))
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("putOficina status %d %s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func putOficinaAtivo(t *testing.T, h http.Handler, id string, ativo bool) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/oficinas/"+id+"/ativo", jsonBody(map[string]any{"ativo": ativo}))
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("putOficinaAtivo status %d %s", rec.Code, rec.Body.String())
	}
}

func getCatalogos(t *testing.T, h http.Handler) (unidades []map[string]any, oficinas []map[string]any) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/catalogos", nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("getCatalogos status %d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Unidades []map[string]any `json:"unidades"`
		Oficinas []map[string]any `json:"oficinas"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body.Unidades, body.Oficinas
}

func getAdminUnidades(t *testing.T, h http.Handler) []map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/unidades", nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("getAdminUnidades status %d %s", rec.Code, rec.Body.String())
	}
	var out []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func getAdminOficinas(t *testing.T, h http.Handler) []map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/oficinas", nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("getAdminOficinas status %d %s", rec.Code, rec.Body.String())
	}
	var out []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestCRUDUnidade(t *testing.T) {
	h := setup(t)

	// Create
	novaUnidade := postUnidade(t, h, "Nova Unidade Teste")
	if novaUnidade["nome"] != "Nova Unidade Teste" || novaUnidade["id"] == "" || novaUnidade["ativo"] != true {
		t.Fatalf("unidade criada inválida: %#v", novaUnidade)
	}

	// Update name
	updatedUnidade := putUnidade(t, h, novaUnidade["id"].(string), "Unidade Teste Atualizada")
	if updatedUnidade["nome"] != "Unidade Teste Atualizada" {
		t.Fatalf("nome da unidade não atualizado: %#v", updatedUnidade)
	}

	// Inactivate
	putUnidadeAtivo(t, h, novaUnidade["id"].(string), false)
	adminUnidades := getAdminUnidades(t, h)
	found := false
	for _, u := range adminUnidades {
		if u["id"] == novaUnidade["id"] {
			if u["ativo"] != false {
				t.Fatalf("unidade não inativada no admin: %#v", u)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("unidade inativada não encontrada na listagem admin")
	}

	// Catalog should omit inactive
	unidades, _ := getCatalogos(t, h)
	for _, u := range unidades {
		if u["id"] == novaUnidade["id"] {
			t.Fatalf("catálogo inclui unidade inativa: %#v", u)
		}
	}

	// Reactivate
	putUnidadeAtivo(t, h, novaUnidade["id"].(string), true)
	unidades, _ = getCatalogos(t, h)
	found = false
	for _, u := range unidades {
		if u["id"] == novaUnidade["id"] {
			if u["ativo"] != true {
				t.Fatalf("unidade não reativada no catálogo: %#v", u)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("unidade reativada não encontrada no catálogo")
	}
}

func TestCRUDOficina(t *testing.T) {
	h := setup(t)

	// Create
	novaOficina := postOficina(t, h, "Nova Oficina Teste")
	if novaOficina["nome"] != "Nova Oficina Teste" || novaOficina["id"] == "" || novaOficina["ativo"] != true {
		t.Fatalf("oficina criada inválida: %#v", novaOficina)
	}

	// Update name
	updatedOficina := putOficina(t, h, novaOficina["id"].(string), "Oficina Teste Atualizada")
	if updatedOficina["nome"] != "Oficina Teste Atualizada" {
		t.Fatalf("nome da oficina não atualizado: %#v", updatedOficina)
	}

	// Inactivate
	putOficinaAtivo(t, h, novaOficina["id"].(string), false)
	adminOficinas := getAdminOficinas(t, h)
	found := false
	for _, o := range adminOficinas {
		if o["id"] == novaOficina["id"] {
			if o["ativo"] != false {
				t.Fatalf("oficina não inativada no admin: %#v", o)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("oficina inativada não encontrada na listagem admin")
	}

	// Catalog should omit inactive
	_, oficinas := getCatalogos(t, h)
	for _, o := range oficinas {
		if o["id"] == novaOficina["id"] {
			t.Fatalf("catálogo inclui oficina inativa: %#v", o)
		}
	}

	// Reactivate
	putOficinaAtivo(t, h, novaOficina["id"].(string), true)
	_, oficinas = getCatalogos(t, h)
	found = false
	for _, o := range oficinas {
		if o["id"] == novaOficina["id"] {
			if o["ativo"] != true {
				t.Fatalf("oficina não reativada no catálogo: %#v", o)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("oficina reativada não encontrada no catálogo")
	}
}

func TestCatalogoIncludesSeedItemsWhenAllActive(t *testing.T) {
	h := setup(t)
	unidades, oficinas := getCatalogos(t, h)

	if len(unidades) != 3 || len(oficinas) != 6 {
		t.Fatalf("Expected 3 units and 6 workshops in initial catalog, got %d units and %d workshops", len(unidades), len(oficinas))
	}

	expectedUnitIDs := map[string]bool{"centro": true, "norte": true, "sul": true}
	for _, u := range unidades {
		if _, ok := expectedUnitIDs[u["id"].(string)]; !ok {
			t.Fatalf("Unexpected unit in catalog: %s", u["id"])
		}
		if u["ativo"] != true {
			t.Fatalf("Seed unit %s is not active", u["id"])
		}
	}

	expectedOficinaIDs := map[string]bool{
		"jiu-jitsu":        true,
		"nacao-esporte":    true,
		"nacao-cultura":    true,
		"acessuas-trabalho": true,
		"karate":           true,
		"teatro":           true,
	}
	for _, o := range oficinas {
		if _, ok := expectedOficinaIDs[o["id"].(string)]; !ok {
			t.Fatalf("Unexpected oficina in catalog: %s", o["id"])
		}
		if o["ativo"] != true {
			t.Fatalf("Seed oficina %s is not active", o["id"])
		}
	}
}
