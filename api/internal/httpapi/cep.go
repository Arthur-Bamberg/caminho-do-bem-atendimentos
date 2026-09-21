package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"unicode"
)

func (s *Server) consultarCEP(w http.ResponseWriter, r *http.Request) {
	cep := soDigitosCEP(r.PathValue("cep"))
	if len(cep) != 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"campo": "cep", "erro": "CEP inválido"})
		return
	}
	base := strings.TrimRight(s.viacepBase, "/")
	url := base + "/ws/" + cep + "/json/"
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"campo": "cep", "erro": "Não foi possível consultar o CEP."})
		return
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"campo": "cep", "erro": "Não foi possível consultar o CEP."})
		return
	}
	defer resp.Body.Close()
	corpo, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil || resp.StatusCode != http.StatusOK {
		writeJSON(w, http.StatusBadGateway, map[string]string{"campo": "cep", "erro": "Não foi possível consultar o CEP."})
		return
	}
	var via struct {
		Erro       json.RawMessage `json:"erro"`
		Logradouro string          `json:"logradouro"`
		Bairro     string          `json:"bairro"`
		Localidade string          `json:"localidade"`
		UF         string          `json:"uf"`
		CEP        string          `json:"cep"`
	}
	if err := json.Unmarshal(corpo, &via); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"campo": "cep", "erro": "Não foi possível consultar o CEP."})
		return
	}
	if viaCEPErro(via.Erro) {
		writeJSON(w, http.StatusNotFound, map[string]string{"campo": "cep", "erro": "CEP não encontrado."})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"logradouro": strings.TrimSpace(via.Logradouro),
		"bairro":     strings.TrimSpace(via.Bairro),
		"cidade":     strings.TrimSpace(via.Localidade),
		"uf":         strings.TrimSpace(via.UF),
		"cep":        soDigitosCEP(via.CEP),
	})
}

func viaCEPErro(raw json.RawMessage) bool {
	s := strings.TrimSpace(string(raw))
	return s == "true" || s == `"true"`
}

func soDigitosCEP(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
