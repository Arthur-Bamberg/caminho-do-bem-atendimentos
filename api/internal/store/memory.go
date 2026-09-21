package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type sessaoMem struct {
	operadorID string
	expira     time.Time
}

type cadastroMem struct {
	cadastro     Cadastro
	atendimentos []Atendimento
}

type Memory struct {
	mu         sync.Mutex
	operadores map[string]Operador
	hashes     map[string]string
	sessoes    map[string]sessaoMem
	cadastros  map[string]cadastroMem
	cpfs       map[string]string
	nucleos    map[string]Nucleo
	unidades   map[string]Unidade
	oficinas   map[string]Oficina
}

func NewMemory() *Memory {
	m := &Memory{
		operadores: make(map[string]Operador),
		hashes:     make(map[string]string),
		sessoes:    make(map[string]sessaoMem),
		cadastros:  make(map[string]cadastroMem),
		cpfs:       make(map[string]string),
		nucleos:    make(map[string]Nucleo),
		unidades:   make(map[string]Unidade),
		oficinas:   make(map[string]Oficina),
	}

	// Seed initial data for units and offices
	initialUnidades := []Unidade{
		{ID: "centro", Nome: "Centro", Ativo: true},
		{ID: "norte", Nome: "Norte", Ativo: true},
		{ID: "sul", Nome: "Sul", Ativo: true},
	}
	for _, u := range initialUnidades {
		m.unidades[u.ID] = u
	}

	initialOficinas := []Oficina{
		{ID: "jiu-jitsu", Nome: "Jiu-jitsu", Ativo: true},
		{ID: "npa-7-12", Nome: "NPA (7 a 12 anos)", Ativo: true},
		{ID: "nacao-esporte", Nome: "Nação Esporte", Ativo: true},
		{ID: "nacao-cultura", Nome: "Nação Cultura", Ativo: true},
		{ID: "acessuas-trabalho", Nome: "Acessuas Trabalho", Ativo: true},
		{ID: "juventude-na-mesa", Nome: "Juventude na Mesa", Ativo: true},
	}
	for _, o := range initialOficinas {
		m.oficinas[o.ID] = o
	}
	return m
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (m *Memory) SeedOperador(email, senha string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.operadores[email] = Operador{ID: "op-seed", Email: email}
	m.hashes[email] = string(hash)
	return nil
}

func (m *Memory) Autenticar(_ context.Context, email, senha string) (Operador, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	op, ok := m.operadores[email]
	if !ok {
		return Operador{}, ErrCredenciaisInvalidas
	}
	if bcrypt.CompareHashAndPassword([]byte(m.hashes[email]), []byte(senha)) != nil {
		return Operador{}, ErrCredenciaisInvalidas
	}
	return op, nil
}

func (m *Memory) CriarSessao(_ context.Context, operadorID, token string, duracao time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessoes[HashToken(token)] = sessaoMem{operadorID: operadorID, expira: time.Now().Add(duracao)}
	return nil
}

func (m *Memory) OperadorDaSessao(_ context.Context, token string) (Operador, error) {
	if token == "" {
		return Operador{}, ErrSessaoInvalida
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessoes[HashToken(token)]
	if !ok || time.Now().After(s.expira) {
		return Operador{}, ErrSessaoInvalida
	}
	for _, op := range m.operadores {
		if op.ID == s.operadorID {
			return op, nil
		}
	}
	return Operador{}, ErrSessaoInvalida
}

func (m *Memory) EncerrarSessao(_ context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessoes, HashToken(token))
	return nil
}

func (m *Memory) CriarCadastro(_ context.Context, in CadastroNovo) (Cadastro, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if in.CPF != "" {
		if _, ok := m.cpfs[in.CPF]; ok {
			return Cadastro{}, ErrCPFDuplicado
		}
	}
	nucleo, err := m.resolverNucleoLocked(in.Nucleo)
	if err != nil {
		return Cadastro{}, err
	}

	if in.UnidadeID != "" {
		u, ok := m.unidades[in.UnidadeID]
		if !ok {
			return Cadastro{}, ErrUnidadeInvalida
		}
		if !u.Ativo {
			return Cadastro{}, ErrUnidadeInativa
		}
	}

	ofs := oficinasMem(m.oficinas, false)
	_, err = ValidarOficinas(ofs, in.Oficinas, false)
	if err != nil {
		return Cadastro{}, err
	}

	id := newID()
	atID := newID()
	in.Nucleo.ResponsavelLegal = nucleo.ResponsavelLegal
	in.Nucleo.WhatsApp = nucleo.WhatsApp
	in.Nucleo.Cesta = nucleo.Cesta
	in.Nucleo.Endereco = nucleo.Endereco
	c := montarCadastro(id, nucleo.ID, atID, in)
	m.cadastros[c.ID] = cadastroMem{cadastro: c, atendimentos: []Atendimento{c.Atendimento}}
	if in.CPF != "" {
		m.cpfs[in.CPF] = id
	}
	return c, nil
}

func (m *Memory) resolverNucleoLocked(in NucleoNovo) (Nucleo, error) {
	if in.ID != "" {
		existente, ok := m.nucleos[in.ID]
		if !ok {
			return Nucleo{}, ErrNucleoNaoEncontrado
		}
		if !NucleoPayloadVazio(in) {
			existente.ResponsavelLegal = in.ResponsavelLegal
			existente.WhatsApp = in.WhatsApp
			existente.Cesta = in.Cesta
			existente.Endereco = in.Endereco
			m.nucleos[in.ID] = existente
			m.propagarNucleoLocked(existente)
		}
		return m.nucleos[in.ID], nil
	}
	n := Nucleo{
		ID:               newID(),
		ResponsavelLegal: in.ResponsavelLegal,
		WhatsApp:         in.WhatsApp,
		Cesta:            in.Cesta,
		Endereco:         in.Endereco,
	}
	m.nucleos[n.ID] = n
	return n, nil
}

func (m *Memory) propagarNucleoLocked(n Nucleo) {
	for _, item := range m.cadastros {
		if item.cadastro.NucleoID == n.ID {
			c := item.cadastro
			c.Nucleo = n
			m.cadastros[c.ID] = cadastroMem{cadastro: c, atendimentos: item.atendimentos}
		}
	}
}

func montarCadastro(id, nucleoID, atID string, in CadastroNovo) Cadastro {
	ofs := in.Oficinas
	if ofs == nil {
		ofs = []string{}
	}
	return Cadastro{
		ID:             id,
		NucleoID:       nucleoID,
		Nome:           in.Nome,
		NomeExibicao:   NomeExibicao(in.Nome),
		CPF:            in.CPF,
		UnidadeID:      in.UnidadeID,
		Oficinas:       ofs,
		Aluno:          EhAluno(ofs),
		Estrangeiro:    in.Estrangeiro,
		PaisOrigem:     in.PaisOrigem,
		DataNascimento: formatarNascimento(in.DataNascimento),
		Profissao:      in.Profissao,
		Nucleo: Nucleo{
			ID:               nucleoID,
			ResponsavelLegal: in.Nucleo.ResponsavelLegal,
			WhatsApp:         in.Nucleo.WhatsApp,
			Cesta:            in.Nucleo.Cesta,
			Endereco:         in.Nucleo.Endereco,
		},
		Atendimento: Atendimento{
			ID:             atID,
			Data:           in.Atendimento.Data.Format("2006-01-02"),
			Relato:         in.Atendimento.Relato,
			ItensEntregues: in.Atendimento.ItensEntregues,
		},
	}
}

func formatarNascimento(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func (m *Memory) AtualizarCadastro(_ context.Context, id string, in CadastroNovo) (Cadastro, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	oldCad, ok := m.cadastros[id]
	if !ok {
		return Cadastro{}, ErrCadastroNaoEncontrado
	}

	if in.CPF != "" {
		if outroID, ok := m.cpfs[in.CPF]; ok && outroID != id {
			return Cadastro{}, ErrCPFDuplicado
		}
	}

	if oldCad.cadastro.CPF != "" && oldCad.cadastro.CPF != in.CPF {
		delete(m.cpfs, oldCad.cadastro.CPF)
	}

	if in.CPF != "" {
		m.cpfs[in.CPF] = id
	}

	if in.UnidadeID != "" {
		u, ok := m.unidades[in.UnidadeID]
		if !ok {
			return Cadastro{}, ErrUnidadeInvalida
		}
		if !u.Ativo && u.ID != oldCad.cadastro.UnidadeID {
			return Cadastro{}, ErrUnidadeInativa
		}
	}

	ofs := oficinasMem(m.oficinas, true)
	_, err := ValidarOficinas(ofs, in.Oficinas, true)
	if err != nil {
		return Cadastro{}, err
	}

	nucleo, err := m.resolverNucleoLocked(in.Nucleo)
	if err != nil {
		return Cadastro{}, err
	}

	atID := newID()
	in.Nucleo.ResponsavelLegal = nucleo.ResponsavelLegal
	in.Nucleo.WhatsApp = nucleo.WhatsApp
	in.Nucleo.Cesta = nucleo.Cesta
	in.Nucleo.Endereco = nucleo.Endereco
	c := montarCadastro(id, oldCad.cadastro.NucleoID, atID, in)
	ats := append(oldCad.atendimentos, c.Atendimento)
	m.cadastros[c.ID] = cadastroMem{cadastro: c, atendimentos: ats}
	return c, nil
}

func (m *Memory) LerCadastro(_ context.Context, id string) (Cadastro, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.cadastros[id]
	if !ok {
		return Cadastro{}, ErrCadastroNaoEncontrado
	}
	return c.cadastro, nil
}

func (m *Memory) Catalogos(ctx context.Context) (Catalogos, error) {
	unidades, err := m.ListarUnidades(ctx, false)
	if err != nil {
		return Catalogos{}, err
	}
	oficinas, err := m.ListarOficinas(ctx, false)
	if err != nil {
		return Catalogos{}, err
	}
	return Catalogos{Unidades: unidades, Oficinas: oficinas}, nil
}

func (m *Memory) ListarUnidades(_ context.Context, incluirInativos bool) ([]Unidade, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Unidade, 0, len(m.unidades))
	for _, u := range m.unidades {
		if incluirInativos || u.Ativo {
			out = append(out, u)
		}
	}
	return out, nil
}

func (m *Memory) CriarUnidade(_ context.Context, nome string) (Unidade, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u := Unidade{ID: MakeSlug(nome), Nome: nome, Ativo: true}
	m.unidades[u.ID] = u
	return u, nil
}

func (m *Memory) AtualizarUnidade(_ context.Context, id, nome string) (Unidade, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.unidades[id]
	if !ok {
		return Unidade{}, ErrUnidadeInvalida
	}
	u.Nome = nome
	m.unidades[id] = u
	return u, nil
}

func (m *Memory) MudarAtivoUnidade(_ context.Context, id string, ativo bool) (Unidade, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.unidades[id]
	if !ok {
		return Unidade{}, ErrUnidadeInvalida
	}
	u.Ativo = ativo
	m.unidades[id] = u
	return u, nil
}

func (m *Memory) GetUnidade(_ context.Context, id string) (Unidade, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.unidades[id]
	if !ok {
		return Unidade{}, ErrUnidadeInvalida
	}
	return u, nil
}

func oficinasMem(oficinas map[string]Oficina, incluirInativos bool) []Oficina {
	out := make([]Oficina, 0, len(oficinas))
	for _, o := range oficinas {
		if incluirInativos || o.Ativo {
			out = append(out, o)
		}
	}
	return out
}

func (m *Memory) ListarOficinas(_ context.Context, incluirInativos bool) ([]Oficina, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Oficina, 0, len(m.oficinas))
	for _, o := range m.oficinas {
		if incluirInativos || o.Ativo {
			out = append(out, o)
		}
	}
	return out, nil
}

func (m *Memory) CriarOficina(_ context.Context, nome string) (Oficina, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o := Oficina{ID: MakeSlug(nome), Nome: nome, Ativo: true}
	m.oficinas[o.ID] = o
	return o, nil
}

func (m *Memory) AtualizarOficina(_ context.Context, id, nome string) (Oficina, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.oficinas[id]
	if !ok {
		return Oficina{}, ErrOficinaInvalida
	}
	o.Nome = nome
	m.oficinas[id] = o
	return o, nil
}

func (m *Memory) MudarAtivoOficina(_ context.Context, id string, ativo bool) (Oficina, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.oficinas[id]
	if !ok {
		return Oficina{}, ErrOficinaInvalida
	}
	o.Ativo = ativo
	m.oficinas[id] = o
	return o, nil
}

func (m *Memory) GetOficina(_ context.Context, id string) (Oficina, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.oficinas[id]
	if !ok {
		return Oficina{}, ErrOficinaInvalida
	}
	return o, nil
}

func (m *Memory) ListarCadastros(_ context.Context) ([]Cadastro, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Cadastro, 0, len(m.cadastros))
	for _, item := range m.cadastros {
		c := item.cadastro
		if n, ok := m.nucleos[item.cadastro.NucleoID]; ok {
			c.Nucleo = n
		}
		out = append(out, c)
	}
	return out, nil
}

func (m *Memory) ConsultarCadastros(ctx context.Context, f ConsultaFiltro) (ConsultaResultado, error) {
	todos, err := m.ListarCadastros(ctx)
	if err != nil {
		return ConsultaResultado{}, err
	}
	filtrados := make([]Cadastro, 0)
	q := strings.ToLower(strings.TrimSpace(f.Q))
	for _, c := range todos {
		if f.UnidadeID != "" && c.UnidadeID != f.UnidadeID {
			continue
		}
		if q != "" && !CadastroCombinaBusca(c, q) {
			continue
		}
		filtrados = append(filtrados, c)
	}
	pagina, por := f.Pagina, f.PorPagina
	if pagina < 1 {
		pagina = 1
	}
	if por < 1 {
		por = 20
	}
	ini := (pagina - 1) * por
	if ini > len(filtrados) {
		ini = len(filtrados)
	}
	fim := ini + por
	if fim > len(filtrados) {
		fim = len(filtrados)
	}
	itens := filtrados[ini:fim]
	if itens == nil {
		itens = []Cadastro{}
	}
	return ConsultaResultado{Itens: itens, Total: len(filtrados), Pagina: pagina, PorPagina: por}, nil
}

func (m *Memory) Indicadores(ctx context.Context, unidadeID string) (Indicadores, error) {
	todos, err := m.ListarCadastros(ctx)
	if err != nil {
		return Indicadores{}, err
	}
	unidades, err := m.ListarUnidades(ctx, true)
	if err != nil {
		return Indicadores{}, err
	}
	oficinas, err := m.ListarOficinas(ctx, true)
	if err != nil {
		return Indicadores{}, err
	}
	return CalcularIndicadores(todos, unidadeID, unidades, oficinas), nil
}

func (m *Memory) ListarAtendimentos(_ context.Context, assistidoID string) ([]Atendimento, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.cadastros[assistidoID]
	if !ok {
		return nil, ErrCadastroNaoEncontrado
	}
	return append([]Atendimento{}, c.atendimentos...), nil
}

func (m *Memory) BuscarNucleos(_ context.Context, q string) ([]NucleoLista, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	q = strings.TrimSpace(q)
	out := []NucleoLista{}
	for _, n := range m.nucleos {
		if q != "" && !strings.Contains(strings.ToLower(n.ResponsavelLegal), q) &&
			!strings.Contains(strings.ToLower(n.WhatsApp), q) &&
			!strings.Contains(strings.ToLower(n.Endereco.Cidade), q) &&
			!strings.Contains(strings.ToLower(n.Endereco.Logradouro), q) {
			continue
		}
		out = append(out, NucleoLista{Nucleo: n, Assistidos: m.contarAssistidosLocked(n.ID)})
	}
	return out, nil
}

func (m *Memory) LerNucleo(_ context.Context, id string) (NucleoLista, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.nucleos[id]
	if !ok {
		return NucleoLista{}, ErrNucleoNaoEncontrado
	}
	return NucleoLista{Nucleo: n, Assistidos: m.contarAssistidosLocked(id)}, nil
}

func (m *Memory) contarAssistidosLocked(nucleoID string) int {
	n := 0
	for _, c := range m.cadastros {
		if c.cadastro.NucleoID == nucleoID {
			n++
		}
	}
	return n
}
