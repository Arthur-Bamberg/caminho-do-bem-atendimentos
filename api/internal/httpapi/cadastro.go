package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Arthur-Bamberg/nacao-assistidos/api/internal/store"
)

var errDataInvalida = errors.New("data inválida")

type cadastroPedido struct {
	Nome           string   `json:"nome"`
	CPF            string   `json:"cpf"`
	Estrangeiro    bool     `json:"estrangeiro"`
	PaisOrigem     string   `json:"pais_origem"`
	DataNascimento string   `json:"data_nascimento"`
	Profissao      string   `json:"profissao"`
	UnidadeID      string   `json:"unidade_id"`
	Oficinas       []string `json:"oficinas"`
	Nucleo         struct {
		ID               string         `json:"id"`
		ResponsavelLegal string         `json:"responsavel_legal"`
		WhatsApp         string         `json:"whatsapp"`
		Cesta            bool           `json:"cesta"`
		Endereco         store.Endereco `json:"endereco"`
	} `json:"nucleo"`
	Atendimento struct {
		Data           string `json:"data"`
		Relato         string `json:"relato"`
		ItensEntregues string `json:"itens_entregues"`
	} `json:"atendimento"`
}

func (s *Server) criarCadastro(w http.ResponseWriter, r *http.Request) {
	in, ok := s.lerPedidoCadastro(w, r)
	if !ok {
		return
	}
	c, err := s.store.CriarCadastro(r.Context(), in)
	s.responderCadastro(w, r, http.StatusCreated, c, err)
}

func (s *Server) atualizarCadastro(w http.ResponseWriter, r *http.Request) {
	in, ok := s.lerPedidoCadastro(w, r)
	if !ok {
		return
	}
	c, err := s.store.AtualizarCadastro(r.Context(), r.PathValue("id"), in)
	s.responderCadastro(w, r, http.StatusOK, c, err)
}

func (s *Server) lerCadastro(w http.ResponseWriter, r *http.Request) {
	c, err := s.store.LerCadastro(r.Context(), r.PathValue("id"))
	s.responderCadastro(w, r, http.StatusOK, c, err)
}

func (s *Server) catalogos(w http.ResponseWriter, r *http.Request) {
	c, err := s.store.Catalogos(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": "falha ao ler catálogos"})
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) buscarNucleos(w http.ResponseWriter, r *http.Request) {
	itens, err := s.store.BuscarNucleos(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": "falha ao buscar núcleos"})
		return
	}
	if itens == nil {
		itens = []store.NucleoLista{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"itens": itens})
}

func (s *Server) lerNucleo(w http.ResponseWriter, r *http.Request) {
	n, err := s.store.LerNucleo(r.Context(), r.PathValue("id"))
	if errors.Is(err, store.ErrNucleoNaoEncontrado) {
		writeJSON(w, http.StatusNotFound, map[string]string{"erro": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": "falha ao ler núcleo"})
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (s *Server) lerPedidoCadastro(w http.ResponseWriter, r *http.Request) (store.CadastroNovo, bool) {
	var body cadastroPedido
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "pedido inválido"})
			return store.CadastroNovo{}, false
		}
	}
	cpf, err := store.NormalizarCPF(body.CPF)
	if errors.Is(err, store.ErrCPFInvalido) {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"campo": "cpf", "erro": "CPF inválido"})
		return store.CadastroNovo{}, false
	}
	if body.UnidadeID != "" {
		u, errU := s.store.GetUnidade(r.Context(), body.UnidadeID)
		if errU != nil || store.ValidarUnidade(u, true) != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"campo": "unidade_id", "erro": "Unidade inválida"})
			return store.CadastroNovo{}, false
		}
	}
	catalogoOficinas, err := s.store.ListarOficinas(r.Context(), true)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": "falha ao ler oficinas"})
		return store.CadastroNovo{}, false
	}
	oficinas, err := store.ValidarOficinas(catalogoOficinas, body.Oficinas, true)
	if errors.Is(err, store.ErrOficinaInvalida) || errors.Is(err, store.ErrOficinaInativa) {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"campo": "oficinas", "erro": "Oficina inválida"})
		return store.CadastroNovo{}, false
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": "falha ao validar oficinas"})
		return store.CadastroNovo{}, false
	}
	data, err := dataAtendimento(body.Atendimento.Data)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"campo": "atendimento.data", "erro": "data inválida"})
		return store.CadastroNovo{}, false
	}
	nasc, err := dataOpcional(body.DataNascimento)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"campo": "data_nascimento", "erro": "data inválida"})
		return store.CadastroNovo{}, false
	}
	paisOrigem := strings.TrimSpace(body.PaisOrigem)
	if !body.Estrangeiro {
		paisOrigem = ""
	}
	hoje := time.Now().In(fusoBrasil())
	return store.CadastroNovo{
		Nome:           strings.TrimSpace(body.Nome),
		CPF:            cpf,
		Estrangeiro:    body.Estrangeiro,
		PaisOrigem:     paisOrigem,
		DataNascimento: nasc,
		Profissao:      store.ProfissaoDoAssistido(nasc, body.Profissao, hoje),
		UnidadeID:      body.UnidadeID,
		Oficinas:       oficinas,
		Nucleo: store.NucleoNovo{
			ID:               strings.TrimSpace(body.Nucleo.ID),
			ResponsavelLegal: strings.TrimSpace(body.Nucleo.ResponsavelLegal),
			WhatsApp:         strings.TrimSpace(body.Nucleo.WhatsApp),
			Cesta:            body.Nucleo.Cesta,
			Endereco: store.Endereco{
				Logradouro:  strings.TrimSpace(body.Nucleo.Endereco.Logradouro),
				Numero:      strings.TrimSpace(body.Nucleo.Endereco.Numero),
				Complemento: strings.TrimSpace(body.Nucleo.Endereco.Complemento),
				Bairro:      strings.TrimSpace(body.Nucleo.Endereco.Bairro),
				Cidade:      strings.TrimSpace(body.Nucleo.Endereco.Cidade),
				UF:          strings.TrimSpace(body.Nucleo.Endereco.UF),
				CEP:         strings.TrimSpace(body.Nucleo.Endereco.CEP),
			},
		},
		Atendimento: store.AtendimentoNovo{
			Data:           data,
			Relato:         strings.TrimSpace(body.Atendimento.Relato),
			ItensEntregues: strings.TrimSpace(body.Atendimento.ItensEntregues),
		},
	}, true
}

func (s *Server) responderCadastro(w http.ResponseWriter, _ *http.Request, okStatus int, c store.Cadastro, err error) {
	if errors.Is(err, store.ErrCPFDuplicado) {
		writeJSON(w, http.StatusConflict, map[string]string{"campo": "cpf", "erro": "CPF já cadastrado"})
		return
	}
	if errors.Is(err, store.ErrCadastroNaoEncontrado) || errors.Is(err, store.ErrNucleoNaoEncontrado) {
		writeJSON(w, http.StatusNotFound, map[string]string{"erro": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": "falha ao gravar cadastro"})
		return
	}
	if c.Oficinas == nil {
		c.Oficinas = []string{}
	}
	writeJSON(w, okStatus, c)
}

func (s *Server) listarCadastros(w http.ResponseWriter, r *http.Request) {
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
	if res.Itens == nil {
		res.Itens = []store.Cadastro{}
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) listarAtendimentos(w http.ResponseWriter, r *http.Request) {
	itens, err := s.store.ListarAtendimentos(r.Context(), r.PathValue("id"))
	if errors.Is(err, store.ErrCadastroNaoEncontrado) {
		writeJSON(w, http.StatusNotFound, map[string]string{"erro": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": "falha ao listar atendimentos"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"itens": itens})
}

func dataOpcional(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	t, err := parseData(s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func dataAtendimento(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	loc := fusoBrasil()
	if s == "" {
		agora := time.Now().In(loc)
		return time.Date(agora.Year(), agora.Month(), agora.Day(), 0, 0, 0, 0, loc), nil
	}
	return parseData(s)
}

func parseData(s string) (time.Time, error) {
	loc := fusoBrasil()
	if t, err := time.ParseInLocation("2006-01-02", s, loc); err == nil && t.Format("2006-01-02") == s {
		return t, nil
	}
	if t, err := time.ParseInLocation("02/01/2006", s, loc); err == nil && t.Format("02/01/2006") == s {
		return t, nil
	}
	return time.Time{}, errDataInvalida
}

func fusoBrasil() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.Local
	}
	return loc
}
