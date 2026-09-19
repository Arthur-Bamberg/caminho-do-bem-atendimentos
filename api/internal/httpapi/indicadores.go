package httpapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Arthur-Bamberg/nacao-assistidos/api/internal/store"
)

func (s *Server) indicadores(w http.ResponseWriter, r *http.Request) {
	uid, ok := unidadeFiltro(w, r, s.store)
	if !ok {
		return
	}
	ind, err := s.store.Indicadores(r.Context(), uid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": "falha ao ler indicadores"})
		return
	}
	writeJSON(w, http.StatusOK, ind)
}

func (s *Server) indicadoresPDF(w http.ResponseWriter, r *http.Request) {
	uid, ok := unidadeFiltro(w, r, s.store)
	if !ok {
		return
	}
	ind, err := s.store.Indicadores(r.Context(), uid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": "falha ao ler indicadores"})
		return
	}
	linhas := []string{
		"Painel de indicadores",
		"filtro_unidade=" + vazioComoTodas(uid),
		fmt.Sprintf("familias=%d", ind.Familias),
		fmt.Sprintf("assistidos=%d", ind.Assistidos),
		fmt.Sprintf("cestas=%d", ind.Cestas),
		"Assistidos por Unidade",
	}
	for _, se := range ind.PorUnidade {
		linhas = append(linhas, fmt.Sprintf("%s=%d", se.ID, se.Quantidade))
	}
	linhas = append(linhas, "Alunos por Oficina")
	for _, se := range ind.PorOficina {
		linhas = append(linhas, fmt.Sprintf("%s=%d", se.ID, se.Quantidade))
	}
	enviarPDF(w, "painel.pdf", linhas)
}

func (s *Server) consultaPDF(w http.ResponseWriter, r *http.Request) {
	pagina, _ := strconv.Atoi(r.URL.Query().Get("pagina"))
	por, _ := strconv.Atoi(r.URL.Query().Get("por_pagina"))
	res, err := s.store.ConsultarCadastros(r.Context(), store.ConsultaFiltro{
		Q:         r.URL.Query().Get("q"),
		UnidadeID: r.URL.Query().Get("unidade_id"),
		Pagina:    pagina,
		PorPagina: por,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": "falha ao listar"})
		return
	}
	linhas := []string{
		"Consulta de Cadastros",
		fmt.Sprintf("total=%d pagina=%d por_pagina=%d", res.Total, res.Pagina, res.PorPagina),
	}
	for _, c := range res.Itens {
		linhas = append(linhas, fmt.Sprintf("%s | %s | %s", c.NomeExibicao, c.UnidadeID, c.CPF))
	}
	enviarPDF(w, "consulta.pdf", linhas)
}

func unidadeFiltro(w http.ResponseWriter, r *http.Request, st store.Store) (string, bool) {
	uid := r.URL.Query().Get("unidade_id")
	if uid != "" {
		if _, err := st.GetUnidade(r.Context(), uid); err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"campo": "unidade_id", "erro": "Unidade inválida"})
			return "", false
		}
	}
	return uid, true
}

func vazioComoTodas(uid string) string {
	if uid == "" {
		return "todas"
	}
	return uid
}

func enviarPDF(w http.ResponseWriter, nome string, linhas []string) {
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="`+nome+`"`)
	_, _ = w.Write(montarPDF(linhas))
}
