package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCadastroAnonimo401(t *testing.T) {
	h := setup(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/assistidos", jsonBody(map[string]any{})))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestCadastroVazioCriaNucleoEAtendimento(t *testing.T) {
	h := setup(t)
	c := postCadastro(t, h, map[string]any{})
	if c["id"] == "" || c["nucleo_id"] == "" {
		t.Fatalf("ids ausentes %#v", c)
	}
	if c["nome"] != "" {
		t.Fatalf("nome %q", c["nome"])
	}
	if c["nome_exibicao"] != "Sem nome" {
		t.Fatalf("exibição %q", c["nome_exibicao"])
	}
	at := c["atendimento"].(map[string]any)
	hoje := time.Now().In(fusoBrasil()).Format("2006-01-02")
	if at["data"] != hoje {
		t.Fatalf("data %q queria %s", at["data"], hoje)
	}
	if at["id"] == "" {
		t.Fatal("atendimento sem id")
	}
}

func TestDoisCadastrosSemCPFNaoFundem(t *testing.T) {
	h := setup(t)
	a := postCadastro(t, h, map[string]any{"nome": "Ana"})
	b := postCadastro(t, h, map[string]any{"nome": "Ana"})
	if a["id"] == b["id"] {
		t.Fatal("ids iguais")
	}
	if a["nucleo_id"] == b["nucleo_id"] {
		t.Fatal("núcleos iguais")
	}
}

func TestCPFDuplicadoEVazioPermitido(t *testing.T) {
	h := setup(t)
	_ = postCadastro(t, h, map[string]any{"cpf": "529.982.247-25"})
	_ = postCadastro(t, h, map[string]any{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/assistidos", jsonBody(map[string]any{"cpf": "52998224725"}))
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["campo"] != "cpf" {
		t.Fatalf("campo %q", body["campo"])
	}
}

func TestRelatoEItensFicamNoAtendimento(t *testing.T) {
	h := setup(t)
	c := postCadastro(t, h, map[string]any{
		"nome": "Bia",
		"atendimento": map[string]any{
			"data":            "2026-09-01",
			"relato":          "chegou hoje",
			"itens_entregues": "leite",
		},
	})
	if c["nome"] == "chegou hoje" {
		t.Fatal("relato gravado na pessoa")
	}
	at := c["atendimento"].(map[string]any)
	if at["data"] != "2026-09-01" || at["relato"] != "chegou hoje" || at["itens_entregues"] != "leite" {
		t.Fatalf("%#v", at)
	}
}

func TestAtendimentoAceitaDataBR(t *testing.T) {
	h := setup(t)
	c := postCadastro(t, h, map[string]any{
		"atendimento": map[string]any{"data": "01/09/2026"},
	})
	at := c["atendimento"].(map[string]any)
	if at["data"] != "2026-09-01" {
		t.Fatalf("data %v", at["data"])
	}
}

func TestListaMinimaSemNome(t *testing.T) {
	h := setup(t)
	_ = postCadastro(t, h, map[string]any{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/assistidos", nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var body struct {
		Itens []map[string]any `json:"itens"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Itens) != 1 || body.Itens[0]["nome_exibicao"] != "Sem nome" {
		t.Fatalf("%#v", body.Itens)
	}
}

func postCadastro(t *testing.T, h http.Handler, body map[string]any) map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/assistidos", jsonBody(body))
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestCadastroRejeitaUnidadeInativaNova(t *testing.T) {
	h := setup(t)

	// Create and inactivate a new unit
	novaUnidade := postUnidade(t, h, "Unidade Inativa Teste")
	putUnidadeAtivo(t, h, novaUnidade["id"].(string), false)

	// Try to register a new assisted with the inactive unit
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/assistidos", jsonBody(map[string]any{"unidade_id": novaUnidade["id"]}))
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Expected status %d for inactive unit, got %d %s", http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["campo"] != "unidade_id" || body["erro"] != "unidade inativa" {
		t.Fatalf("Expected validation error for inactive unit, got %#v", body)
	}
}

func TestCadastroAceitaUnidadeInativaExistente(t *testing.T) {
	h := setup(t)

	// Create a new assisted with an active unit
	ativaUnidade := postUnidade(t, h, "Unidade Ativa Teste")
	c := postCadastro(t, h, map[string]any{"unidade_id": ativaUnidade["id"]})

	// Inactivate the unit
	putUnidadeAtivo(t, h, ativaUnidade["id"].(string), false)

	// Update the existing assisted with the now inactive unit - should be accepted
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/assistidos/"+c["id"].(string), jsonBody(map[string]any{"unidade_id": ativaUnidade["id"]}))
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status %d for existing assisted with inactive unit, got %d %s", http.StatusOK, rec.Code, rec.Body.String())
	}
}

func TestCadastroRejeitaOficinaInativaNova(t *testing.T) {
	h := setup(t)

	// Create and inactivate a new workshop
	novaOficina := postOficina(t, h, "Oficina Inativa Teste")
	putOficinaAtivo(t, h, novaOficina["id"].(string), false)

	// Try to register a new assisted with the inactive workshop
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/assistidos", jsonBody(map[string]any{"oficinas": []string{novaOficina["id"].(string)}}))
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Expected status %d for inactive workshop, got %d %s", http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["campo"] != "oficinas" || body["erro"] != "oficina inativa" {
		t.Fatalf("Expected validation error for inactive workshop, got %#v", body)
	}
}

func TestCadastroAceitaOficinaInativaExistente(t *testing.T) {
	h := setup(t)

	// Create a new assisted with an active workshop
	ativaOficina := postOficina(t, h, "Oficina Ativa Teste")
	c := postCadastro(t, h, map[string]any{"oficinas": []string{ativaOficina["id"].(string)}})

	// Inactivate the workshop
	putOficinaAtivo(t, h, ativaOficina["id"].(string), false)

	// Update the existing assisted with the now inactive workshop - should be accepted
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/assistidos/"+c["id"].(string), jsonBody(map[string]any{"oficinas": []string{ativaOficina["id"].(string)}}))
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status %d for existing assisted with inactive workshop, got %d %s", http.StatusOK, rec.Code, rec.Body.String())
	}
}

func TestEstrangeiroEPaisOrigem(t *testing.T) {
	h := setup(t)

	// Case 1: estrangeiro = true, pais_origem provided
	c1 := postCadastro(t, h, map[string]any{
		"estrangeiro": true,
		"pais_origem": "Paraguai",
	})
	if c1["estrangeiro"] != true || c1["pais_origem"] != "Paraguai" {
		t.Fatalf("Case 1 failed: %#v", c1)
	}

	// Case 2: estrangeiro = true, pais_origem empty
	c2 := postCadastro(t, h, map[string]any{
		"estrangeiro": true,
		"pais_origem": "",
	})
	if c2["estrangeiro"] != true || c2["pais_origem"] != "" {
		t.Fatalf("Case 2 failed: %#v", c2)
	}

	// Case 3: estrangeiro = false, pais_origem provided (should be ignored/emptied)
	c3 := postCadastro(t, h, map[string]any{
		"estrangeiro": false,
		"pais_origem": "Bolívia", // This should be ignored
	})
	if c3["estrangeiro"] != false || c3["pais_origem"] != "" {
		t.Fatalf("Case 3 failed: %#v", c3)
	}

	// Case 4: estrangeiro = false, pais_origem empty
	c4 := postCadastro(t, h, map[string]any{
		"estrangeiro": false,
		"pais_origem": "",
	})
	if c4["estrangeiro"] != false || c4["pais_origem"] != "" {
		t.Fatalf("Case 4 failed: %#v", c4)
	}

	// Test update: set estrangeiro to false, check if pais_origem is emptied
	updatedC1 := putCadastro(t, h, c1["id"].(string), map[string]any{
		"estrangeiro": false,
		"pais_origem": "Argentina", // Should be emptied
	})
	if updatedC1["estrangeiro"] != false || updatedC1["pais_origem"] != "" {
		t.Fatalf("Update case failed: %#v", updatedC1)
	}

	// Test update: set estrangeiro to true, and provide pais_origem
	updatedC2 := putCadastro(t, h, c2["id"].(string), map[string]any{
		"estrangeiro": true,
		"pais_origem": "Chile",
	})
	if updatedC2["estrangeiro"] != true || updatedC2["pais_origem"] != "Chile" {
		t.Fatalf("Update case 2 failed: %#v", updatedC2)
	}
}

func TestDataNascimentoEProfissao(t *testing.T) {
	h := setup(t)
	hoje := time.Now().In(fusoBrasil())
	adulto := hoje.AddDate(-20, 0, 0).Format("2006-01-02")
	menor := hoje.AddDate(-10, 0, 0).Format("2006-01-02")
	maioridade := hoje.AddDate(-18, 0, 0).Format("2006-01-02")

	vazio := postCadastro(t, h, map[string]any{"profissao": "pedreiro"})
	if vazio["data_nascimento"] != "" || vazio["profissao"] != "" {
		t.Fatalf("sem nascimento não grava profissão: %#v", vazio)
	}

	crianca := postCadastro(t, h, map[string]any{
		"data_nascimento": menor,
		"profissao":       "pedreiro",
	})
	if crianca["data_nascimento"] != menor || crianca["profissao"] != "" {
		t.Fatalf("menor de 18 não grava profissão: %#v", crianca)
	}

	exato := postCadastro(t, h, map[string]any{
		"data_nascimento": maioridade,
		"profissao":       "costureira",
	})
	if exato["data_nascimento"] != maioridade || exato["profissao"] != "costureira" {
		t.Fatalf("18 anos grava profissão: %#v", exato)
	}

	adultoC := postCadastro(t, h, map[string]any{
		"data_nascimento": adulto,
		"profissao":       "pedreiro",
	})
	if adultoC["data_nascimento"] != adulto || adultoC["profissao"] != "pedreiro" {
		t.Fatalf("adulto grava profissão: %#v", adultoC)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/assistidos", jsonBody(map[string]any{"data_nascimento": "31/02/2000"}))
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("data inválida status %d %s", rec.Code, rec.Body.String())
	}

	depois := putCadastro(t, h, adultoC["id"].(string), map[string]any{
		"data_nascimento": menor,
		"profissao":       "pedreiro",
	})
	if depois["profissao"] != "" {
		t.Fatalf("atualizar para menor esvazia profissão: %#v", depois)
	}
}
