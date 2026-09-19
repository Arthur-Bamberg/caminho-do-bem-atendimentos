package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestIndicadoresAnonimo401(t *testing.T) {
	h := setup(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/indicadores", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestIndicadoresVazioZeros(t *testing.T) {
	h := setup(t)
	ind := getIndicadores(t, h, "")
	if ind.Familias != 0 || ind.Assistidos != 0 || ind.Cestas != 0 {
		t.Fatalf("%#v", ind)
	}
	if len(ind.PorUnidade) != 3 || len(ind.PorOficina) != 6 {
		t.Fatalf("séries %#v %#v", ind.PorUnidade, ind.PorOficina)
	}
	for _, s := range append(ind.PorUnidade, ind.PorOficina...) {
		if s.Quantidade != 0 || s.Nome == "" {
			t.Fatalf("%#v", s)
		}
	}
}

func TestIndicadoresIrmaosMesmoNucleo(t *testing.T) {
	h := setup(t)
	a := postCadastro(t, h, map[string]any{
		"nome":       "Irmão 1",
		"unidade_id": "centro",
		"nucleo": map[string]any{
			"responsavel_legal": "Maria",
			"cesta":             true,
		},
	})
	nucleoID := a["nucleo"].(map[string]any)["id"].(string)
	_ = postCadastro(t, h, map[string]any{
		"nome":       "Irmão 2",
		"unidade_id": "norte",
		"nucleo":     map[string]any{"id": nucleoID},
	})
	ind := getIndicadores(t, h, "")
	if ind.Familias != 1 || ind.Assistidos != 2 || ind.Cestas != 1 {
		t.Fatalf("todas %#v", ind)
	}
	centro := getIndicadores(t, h, "centro")
	if centro.Familias != 1 || centro.Assistidos != 1 || centro.Cestas != 1 {
		t.Fatalf("centro %#v", centro)
	}
	if serieQtd(centro.PorUnidade, "centro") != 1 || serieQtd(centro.PorUnidade, "norte") != 0 {
		t.Fatalf("por unidade %#v", centro.PorUnidade)
	}
}

func TestIndicadoresAlunosSoOficinaVigente(t *testing.T) {
	h := setup(t)
	_ = postCadastro(t, h, map[string]any{"nome": "Só cesta"})
	aluno := postCadastro(t, h, map[string]any{
		"nome":     "Aluno",
		"oficinas": []string{"jiu-jitsu", "nacao-esporte"},
	})
	ind := getIndicadores(t, h, "")
	if serieQtd(ind.PorOficina, "jiu-jitsu") != 1 || serieQtd(ind.PorOficina, "nacao-esporte") != 1 {
		t.Fatalf("inicial %#v", ind.PorOficina)
	}
	_ = putCadastro(t, h, aluno["id"].(string), map[string]any{
		"nome":     "Aluno",
		"oficinas": []string{"npa-7-12"},
		"nucleo":   map[string]any{},
	})
	depois := getIndicadores(t, h, "")
	if serieQtd(depois.PorOficina, "jiu-jitsu") != 0 || serieQtd(depois.PorOficina, "npa-7-12") != 1 {
		t.Fatalf("vigente %#v", depois.PorOficina)
	}
}

type indicadoresJSON struct {
	UnidadeID  string `json:"unidade_id"`
	Familias   int    `json:"familias"`
	Assistidos int    `json:"assistidos"`
	Cestas     int    `json:"cestas"`
	PorUnidade []struct {
		ID         string `json:"id"`
		Nome       string `json:"nome"`
		Quantidade int    `json:"quantidade"`
	} `json:"por_unidade"`
	PorOficina []struct {
		ID         string `json:"id"`
		Nome       string `json:"nome"`
		Quantidade int    `json:"quantidade"`
	} `json:"por_oficina"`
}

func getIndicadores(t *testing.T, h http.Handler, unidade string) indicadoresJSON {
	t.Helper()
	path := "/api/indicadores"
	if unidade != "" {
		path += "?unidade_id=" + unidade
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	var ind indicadoresJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &ind); err != nil {
		t.Fatal(err)
	}
	return ind
}

func serieQtd(ss []struct {
	ID         string `json:"id"`
	Nome       string `json:"nome"`
	Quantidade int    `json:"quantidade"`
}, id string) int {
	for _, s := range ss {
		if s.ID == id {
			return s.Quantidade
		}
	}
	return -1
}

func TestPDFIndicadoresBateComJSON(t *testing.T) {
	h := setup(t)
	_ = postCadastro(t, h, map[string]any{
		"nome":       "Ana",
		"unidade_id": "centro",
		"nucleo":     map[string]any{"cesta": true},
	})
	ind := getIndicadores(t, h, "centro")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/indicadores/pdf?unidade_id=centro", nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "application/pdf") {
		t.Fatalf("tipo %s", rec.Header().Get("Content-Type"))
	}
	txt := rec.Body.String()
	if !strings.Contains(txt, "%PDF") {
		t.Fatal("não é PDF")
	}
	if !strings.Contains(txt, "familias="+strconv.Itoa(ind.Familias)) ||
		!strings.Contains(txt, "assistidos="+strconv.Itoa(ind.Assistidos)) ||
		!strings.Contains(txt, "cestas="+strconv.Itoa(ind.Cestas)) {
		t.Fatalf("texto %s", txt)
	}
}

func TestPDFConsultaBateComLinhas(t *testing.T) {
	h := setup(t)
	c := postCadastro(t, h, map[string]any{"nome": "Maria PDF"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/assistidos/pdf?q=Maria", nil)
	req.AddCookie(loginCookie(t, h))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	txt := rec.Body.String()
	if !strings.Contains(txt, "Maria PDF") {
		t.Fatalf("linha ausente %s", txt)
	}
	if !strings.Contains(txt, c["id"].(string)[:8]) && !strings.Contains(txt, "Maria PDF") {
		t.Fatal("cadastro não listado")
	}
}

func TestPDFAnonimo401(t *testing.T) {
	h := setup(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/indicadores/pdf", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/assistidos/pdf", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("consulta pdf %d", rec.Code)
	}
}

