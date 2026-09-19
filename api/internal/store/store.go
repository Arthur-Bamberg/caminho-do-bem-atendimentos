package store

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"
)

var (
	ErrCredenciaisInvalidas = errors.New("credenciais inválidas")
	ErrSessaoInvalida       = errors.New("sessão inválida")
	ErrCPFDuplicado         = errors.New("CPF já cadastrado")
	ErrCPFInvalido          = errors.New("CPF inválido")
	ErrUnidadeInvalida      = errors.New("Unidade inválida")
	ErrOficinaInvalida      = errors.New("Oficina inválida")
	ErrCadastroNaoEncontrado = errors.New("Cadastro não encontrado")
	ErrNucleoNaoEncontrado   = errors.New("Núcleo não encontrado")
)

type Operador struct {
	ID    string
	Email string
}

type AtendimentoNovo struct {
	Data           time.Time
	Relato         string
	ItensEntregues string
}

type CadastroNovo struct {
	Nome        string
	CPF         string
	UnidadeID   string
	Oficinas    []string
	Nucleo      NucleoNovo
	Estrangeiro bool
	PaisOrigem  string
	Atendimento AtendimentoNovo
}

type NucleoNovo struct {
	ID               string
	ResponsavelLegal string
	WhatsApp         string
	Cesta            bool
	Endereco         Endereco
}

type Endereco struct {
	Logradouro  string `json:"logradouro"`
	Numero      string `json:"numero"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Cidade      string `json:"cidade"`
	UF          string `json:"uf"`
	CEP         string `json:"cep"`
}

type Nucleo struct {
	ID               string   `json:"id"`
	ResponsavelLegal string   `json:"responsavel_legal"`
	WhatsApp         string   `json:"whatsapp"`
	Cesta            bool     `json:"cesta"`
	Endereco         Endereco `json:"endereco"`
}

type NucleoLista struct {
	Nucleo
	Assistidos int `json:"assistidos"`
}

type Atendimento struct {
	ID             string `json:"id"`
	Data           string `json:"data"`
	Relato         string `json:"relato"`
	ItensEntregues string `json:"itens_entregues"`
}

type Cadastro struct {
	ID           string      `json:"id"`
	NucleoID     string      `json:"nucleo_id"`
	Nome         string      `json:"nome"`
	NomeExibicao string      `json:"nome_exibicao"`
	CPF          string      `json:"cpf"`
	UnidadeID    string      `json:"unidade_id"`
	Oficinas     []string    `json:"oficinas"`
	Aluno        bool        `json:"aluno"`
	Estrangeiro  bool        `json:"estrangeiro"`
	PaisOrigem   string      `json:"pais_origem"`
	Nucleo       Nucleo      `json:"nucleo"`
	Atendimento  Atendimento `json:"atendimento"`
}

type ConsultaFiltro struct {
	Q          string
	UnidadeID  string
	Pagina     int
	PorPagina  int
}

type ConsultaResultado struct {
	Itens     []Cadastro `json:"itens"`
	Total     int        `json:"total"`
	Pagina    int        `json:"pagina"`
	PorPagina int        `json:"por_pagina"`
}

type Store interface {
	Autenticar(ctx context.Context, email, senha string) (Operador, error)
	CriarSessao(ctx context.Context, operadorID, token string, duracao time.Duration) error
	OperadorDaSessao(ctx context.Context, token string) (Operador, error)
	EncerrarSessao(ctx context.Context, token string) error
	CriarCadastro(ctx context.Context, in CadastroNovo) (Cadastro, error)
	AtualizarCadastro(ctx context.Context, id string, in CadastroNovo) (Cadastro, error)
	LerCadastro(ctx context.Context, id string) (Cadastro, error)
	ListarCadastros(ctx context.Context) ([]Cadastro, error)
	ConsultarCadastros(ctx context.Context, f ConsultaFiltro) (ConsultaResultado, error)
	Indicadores(ctx context.Context, unidadeID string) (Indicadores, error)
	ListarAtendimentos(ctx context.Context, assistidoID string) ([]Atendimento, error)
	Catalogos(ctx context.Context) (Catalogos, error)
	BuscarNucleos(ctx context.Context, q string) ([]NucleoLista, error)
	LerNucleo(ctx context.Context, id string) (NucleoLista, error)

	ListarUnidades(ctx context.Context, incluirInativos bool) ([]Unidade, error)
	CriarUnidade(ctx context.Context, nome string) (Unidade, error)
	AtualizarUnidade(ctx context.Context, id, nome string) (Unidade, error)
	MudarAtivoUnidade(ctx context.Context, id string, ativo bool) (Unidade, error)
	GetUnidade(ctx context.Context, id string) (Unidade, error)

	ListarOficinas(ctx context.Context, incluirInativos bool) ([]Oficina, error)
	CriarOficina(ctx context.Context, nome string) (Oficina, error)
	AtualizarOficina(ctx context.Context, id, nome string) (Oficina, error)
	MudarAtivoOficina(ctx context.Context, id string, ativo bool) (Oficina, error)
	GetOficina(ctx context.Context, id string) (Oficina, error)
}

func NomeExibicao(nome string) string {
	if strings.TrimSpace(nome) == "" {
		return "Sem nome"
	}
	return nome
}

func NormalizarCPF(cpf string) (string, error) {
	var b strings.Builder
	for _, r := range cpf {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	n := b.String()
	if n == "" {
		return "", nil
	}
	if len(n) != 11 {
		return "", ErrCPFInvalido
	}
	return n, nil
}
