package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Postgres{pool: pool}, nil
}

func (s *Postgres) Close() {
	s.pool.Close()
}

func (s *Postgres) Migrate(ctx context.Context, sql string) error {
	_, err := s.pool.Exec(ctx, sql)
	return err
}

func (s *Postgres) SeedOperador(ctx context.Context, email, senha string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	email = strings.TrimSpace(strings.ToLower(email))
	_, err = s.pool.Exec(ctx, `
		INSERT INTO operadores (email, senha_hash)
		VALUES ($1, $2)
		ON CONFLICT (email) DO UPDATE SET senha_hash = EXCLUDED.senha_hash
	`, email, string(hash))
	return err
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *Postgres) Autenticar(ctx context.Context, email, senha string) (Operador, error) {
	var op Operador
	var hash string
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, email, senha_hash FROM operadores WHERE email = $1
	`, email).Scan(&op.ID, &op.Email, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return Operador{}, ErrCredenciaisInvalidas
	}
	if err != nil {
		return Operador{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha)) != nil {
		return Operador{}, ErrCredenciaisInvalidas
	}
	return op, nil
}

func (s *Postgres) CriarSessao(ctx context.Context, operadorID, token string, duracao time.Duration) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessoes (operador_id, token_hash, expira_em)
		VALUES ($1, $2, $3)
	`, operadorID, HashToken(token), time.Now().Add(duracao))
	return err
}

func (s *Postgres) OperadorDaSessao(ctx context.Context, token string) (Operador, error) {
	if token == "" {
		return Operador{}, ErrSessaoInvalida
	}
	var op Operador
	err := s.pool.QueryRow(ctx, `
		SELECT o.id::text, o.email
		FROM sessoes s
		JOIN operadores o ON o.id = s.operador_id
		WHERE s.token_hash = $1 AND s.expira_em > now()
	`, HashToken(token)).Scan(&op.ID, &op.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return Operador{}, ErrSessaoInvalida
	}
	if err != nil {
		return Operador{}, err
	}
	return op, nil
}

func (s *Postgres) EncerrarSessao(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	_, err := s.pool.Exec(ctx, `DELETE FROM sessoes WHERE token_hash = $1`, HashToken(token))
	return err
}

func (s *Postgres) CriarCadastro(ctx context.Context, in CadastroNovo) (Cadastro, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Cadastro{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var nucleoID string
	if in.Nucleo.ID != "" {
		var existe bool
		err := tx.QueryRow(ctx, `SELECT true FROM nucleos_familiares WHERE id = $1`, in.Nucleo.ID).Scan(&existe)
		if errors.Is(err, pgx.ErrNoRows) {
			return Cadastro{}, ErrNucleoNaoEncontrado
		}
		if err != nil {
			return Cadastro{}, err
		}
		nucleoID = in.Nucleo.ID
		if !NucleoPayloadVazio(in.Nucleo) {
			if _, err := tx.Exec(ctx, `
				UPDATE nucleos_familiares SET
					responsavel_legal = $2, whatsapp = $3, cesta = $4,
					logradouro = $5, numero = $6, complemento = $7, bairro = $8, cidade = $9, uf = $10, cep = $11
				WHERE id = $1
			`, nucleoID, in.Nucleo.ResponsavelLegal, in.Nucleo.WhatsApp, in.Nucleo.Cesta,
				in.Nucleo.Endereco.Logradouro, in.Nucleo.Endereco.Numero, in.Nucleo.Endereco.Complemento,
				in.Nucleo.Endereco.Bairro, in.Nucleo.Endereco.Cidade, in.Nucleo.Endereco.UF, in.Nucleo.Endereco.CEP,
			); err != nil {
				return Cadastro{}, err
			}
		} else {
			n, err := lerNucleoTx(ctx, tx, nucleoID)
			if err != nil {
				return Cadastro{}, err
			}
			in.Nucleo.ResponsavelLegal = n.ResponsavelLegal
			in.Nucleo.WhatsApp = n.WhatsApp
			in.Nucleo.Cesta = n.Cesta
			in.Nucleo.Endereco = n.Endereco
		}
	} else if err := tx.QueryRow(ctx, `
		INSERT INTO nucleos_familiares (
			responsavel_legal, whatsapp, cesta, logradouro, numero, complemento, bairro, cidade, uf, cep
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id::text
	`, in.Nucleo.ResponsavelLegal, in.Nucleo.WhatsApp, in.Nucleo.Cesta,
		in.Nucleo.Endereco.Logradouro, in.Nucleo.Endereco.Numero, in.Nucleo.Endereco.Complemento,
		in.Nucleo.Endereco.Bairro, in.Nucleo.Endereco.Cidade, in.Nucleo.Endereco.UF, in.Nucleo.Endereco.CEP,
	).Scan(&nucleoID); err != nil {
		return Cadastro{}, err
	}
	var unidade any
	if in.UnidadeID == "" {
		unidade = nil
	} else {
		unidade = in.UnidadeID
	}
	var assistidoID string
	err = tx.QueryRow(ctx, `
		INSERT INTO assistidos (nucleo_id, nome, cpf, unidade_id, estrangeiro, pais_origem)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text
	`, nucleoID, in.Nome, in.CPF, unidade, in.Estrangeiro, in.PaisOrigem).Scan(&assistidoID)
	if err != nil {
		if strings.Contains(err.Error(), "assistidos_cpf_unico") {
			return Cadastro{}, ErrCPFDuplicado
		}
		return Cadastro{}, err
	}
	if err := gravarOficinas(ctx, tx, assistidoID, in.Oficinas); err != nil {
		return Cadastro{}, err
	}
	var atID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO atendimentos (assistido_id, data, relato, itens_entregues)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text
	`, assistidoID, in.Atendimento.Data, in.Atendimento.Relato, in.Atendimento.ItensEntregues).Scan(&atID); err != nil {
		return Cadastro{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Cadastro{}, err
	}
	return montarCadastro(assistidoID, nucleoID, atID, in), nil
}

func gravarOficinas(ctx context.Context, tx pgx.Tx, assistidoID string, oficinas []string) error {
	if _, err := tx.Exec(ctx, `DELETE FROM assistido_oficinas WHERE assistido_id = $1`, assistidoID); err != nil {
		return err
	}
	for _, of := range oficinas {
		if _, err := tx.Exec(ctx, `INSERT INTO assistido_oficinas (assistido_id, oficin-id) VALUES ($1, $2)`, assistidoID, of); err != nil {
			return err
		}
	}
	return nil
}

func (s *Postgres) AtualizarCadastro(ctx context.Context, id string, in CadastroNovo) (Cadastro, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Cadastro{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var nucleoID string
	err = tx.QueryRow(ctx, `SELECT nucleo_id::text FROM assistidos WHERE id = $1`, id).Scan(&nucleoID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cadastro{}, ErrCadastroNaoEncontrado
	}
	if err != nil {
		return Cadastro{}, err
	}
	var unidade any
	if in.UnidadeID == "" {
		unidade = nil
	} else {
		unidade = in.UnidadeID
	}
	tag, err := tx.Exec(ctx, `
		UPDATE assistidos SET nome = $2, cpf = $3, unidade_id = $4, estrangeiro = $5, pais_origem = $6 WHERE id = $1
	`, id, in.Nome, in.CPF, unidade, in.Estrangeiro, in.PaisOrigem)
	if err != nil {
		if strings.Contains(err.Error(), "assistidos_cpf_unico") {
			return Cadastro{}, ErrCPFDuplicado
		}
		return Cadastro{}, err
	}
	if tag.RowsAffected() == 0 {
		return Cadastro{}, ErrCadastroNaoEncontrado
	}
	if _, err := tx.Exec(ctx, `
		UPDATE nucleos_familiares SET
			responsavel_legal = $2, whatsapp = $3, cesta = $4,
			logradouro = $5, numero = $6, complemento = $7, bairro = $8, cidade = $9, uf = $10, cep = $11
		WHERE id = $1
	`, nucleoID, in.Nucleo.ResponsavelLegal, in.Nucleo.WhatsApp, in.Nucleo.Cesta,
		in.Nucleo.Endereco.Logradouro, in.Nucleo.Endereco.Numero, in.Nucleo.Endereco.Complemento,
		in.Nucleo.Endereco.Bairro, in.Nucleo.Endereco.Cidade, in.Nucleo.Endereco.UF, in.Nucleo.Endereco.CEP,
	); err != nil {
		return Cadastro{}, err
	}
	if err := gravarOficinas(ctx, tx, id, in.Oficinas); err != nil {
		return Cadastro{}, err
	}
	var atID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO atendimentos (assistido_id, data, relato, itens_entregues)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text
	`, id, in.Atendimento.Data, in.Atendimento.Relato, in.Atendimento.ItensEntregues).Scan(&atID); err != nil {
		return Cadastro{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Cadastro{}, err
	}
	return montarCadastro(id, nucleoID, atID, in), nil
}

func (s *Postgres) LerCadastro(ctx context.Context, id string) (Cadastro, error) {
	var c Cadastro
	var unidadeID sql.NullString
	var nucleoID string
	var estrangeiro bool
	var paisOrigem string

	err := s.pool.QueryRow(ctx, `
		SELECT a.id::text, a.nucleo_id::text, a.nome, a.cpf, a.unidade_id,
		       a.estrangeiro, a.pais_origem,
		       t.id::text, to_char(t.data, 'YYYY-MM-DD'), t.relato, t.itens_entregues
		FROM assistidos a
		JOIN LATERAL (
			SELECT * FROM atendimentos WHERE assistido_id = a.id ORDER BY data DESC, id DESC LIMIT 1
		) t ON true
		WHERE a.id = $1
	`, id).Scan(&c.ID, &nucleoID, &c.Nome, &c.CPF, &unidadeID,
		&estrangeiro, &paisOrigem,
		&c.Atendimento.ID, &c.Atendimento.Data, &c.Atendimento.Relato, &c.Atendimento.ItensEntregues,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cadastro{}, ErrCadastroNaoEncontrado
	}
	if err != nil {
		return Cadastro{}, err
	}
	c.NucleoID = nucleoID
	c.UnidadeID = unidadeID.String
	c.Estrangeiro = estrangeiro
	c.PaisOrigem = paisOrigem

	ofs, err := s.oficinasDoAssistido(ctx, c.ID)
	if err != nil {
		return Cadastro{}, err
	}
	c.Oficinas = ofs
	c.Aluno = EhAluno(ofs)
	c.NomeExibicao = NomeExibicao(c.Nome)

	n, err := s.LerNucleo(ctx, nucleoID)
	if err != nil {
		return Cadastro{}, err
	}
	c.Nucleo = n.Nucleo
	return c, nil
}

func (s *Postgres) Catalogos(ctx context.Context) (Catalogos, error) {
	unidades, err := s.ListarUnidades(ctx, false)
	if err != nil {
		return Catalogos{}, err
	}
	oficinas, err := s.ListarOficinas(ctx, false)
	if err != nil {
		return Catalogos{}, err
	}
	return Catalogos{Unidades: unidades, Oficinas: oficinas}, nil
}

func (s *Postgres) ListarCadastros(ctx context.Context) ([]Cadastro, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id::text, a.nucleo_id::text, a.nome, a.cpf, COALESCE(a.unidade_id, ''),
		       a.estrangeiro, a.pais_origem,
		       t.id::text, to_char(t.data, 'YYYY-MM-DD'), t.relato, t.itens_entregues,
		       n.responsavel_legal, n.whatsapp, n.cesta,
		       n.logradouro, n.numero, n.complemento, n.bairro, n.cidade, n.uf, n.cep
		FROM assistidos a
		JOIN nucleos_familiares n ON n.id = a.nucleo_id
		JOIN LATERAL (
			SELECT * FROM atendimentos WHERE assistido_id = a.id ORDER BY data DESC, id DESC LIMIT 1
		) t ON true
		ORDER BY a.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Cadastro
	for rows.Next() {
		var c Cadastro
		var unidadeID sql.NullString
		var estrangeiro bool
		var paisOrigem string
		if err := rows.Scan(&c.ID, &c.NucleoID, &c.Nome, &c.CPF, &unidadeID,
			&estrangeiro, &paisOrigem,
			&c.Atendimento.ID, &c.Atendimento.Data, &c.Atendimento.Relato, &c.Atendimento.ItensEntregues,
			&c.Nucleo.ResponsavelLegal, &c.Nucleo.WhatsApp, &c.Nucleo.Cesta,
			&c.Nucleo.Endereco.Logradouro, &c.Nucleo.Endereco.Numero, &c.Nucleo.Endereco.Complemento,
			&c.Nucleo.Endereco.Bairro, &c.Nucleo.Endereco.Cidade, &c.Nucleo.Endereco.UF, &c.Nucleo.Endereco.CEP,
		); err != nil {
			return nil, err
		}
		c.UnidadeID = unidadeID.String
		c.Estrangeiro = estrangeiro
		c.PaisOrigem = paisOrigem
		ofs, err := s.oficinasDoAssistido(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		c.Oficinas = ofs
		c.Aluno = EhAluno(ofs)
		c.NomeExibicao = NomeExibicao(c.Nome)
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Postgres) ConsultarCadastros(ctx context.Context, f ConsultaFiltro) (ConsultaResultado, error) {
	todos, err := s.ListarCadastros(ctx)
	if err != nil {
		return ConsultaResultado{}, err
	}
	filtrados := make([]Cadastro, 0)
	q := strings.ToLower(strings.TrimSpace(f.Q))
	for _, c := range todos {
		if f.UnidadeID != "" && c.UnidadeID != f.UnidadeID {
			continue
		}
		if q != "" {
			blob := strings.ToLower(c.Nome + " " + c.NomeExibicao + " " + c.CPF + " " + c.Nucleo.ResponsavelLegal)
			if !strings.Contains(blob, q) {
				continue
			}
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

func (s *Postgres) Indicadores(ctx context.Context, unidadeID string) (Indicadores, error) {
	unidades, err := s.ListarUnidades(ctx, true)
	if err != nil {
		return Indicadores{}, err
	}
	oficinas, err := s.ListarOficinas(ctx, true)
	if err != nil {
		return Indicadores{}, err
	}
	todos, err := s.ListarCadastros(ctx)
	if err != nil {
		return Indicadores{}, err
	}
	return CalcularIndicadores(todos, unidadeID, unidades, oficinas), nil
}

func (s *Postgres) ListarAtendimentos(ctx context.Context, assistidoID string) ([]Atendimento, error) {
	if _, err := s.LerCadastro(ctx, assistidoID); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, to_char(data, 'YYYY-MM-DD'), relato, itens_entregues
		FROM atendimentos WHERE assistido_id = $1 ORDER BY data, id
	`, assistidoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Atendimento{}
	for rows.Next() {
		var a Atendimento
		if err := rows.Scan(&a.ID, &a.Data, &a.Relato, &a.ItensEntregues); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Postgres) oficinasDoAssistido(ctx context.Context, assistidoID string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `SELECT oficin-id FROM assistido_oficinas WHERE assistido_id = $1 ORDER BY oficin-id`, assistidoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func lerNucleoTx(ctx context.Context, tx pgx.Tx, id string) (Nucleo, error) {
	var n Nucleo
	n.ID = id
	err := tx.QueryRow(ctx, `
		SELECT responsavel_legal, whatsapp, cesta, logradouro, numero, complemento, bairro, cidade, uf, cep
		FROM nucleos_familiares WHERE id = $1
	`, id).Scan(&n.ResponsavelLegal, &n.WhatsApp, &n.Cesta,
		&n.Endereco.Logradouro, &n.Endereco.Numero, &n.Endereco.Complemento,
		&n.Endereco.Bairro, &n.Endereco.Cidade, &n.Endereco.UF, &n.Endereco.CEP)
	if errors.Is(err, pgx.ErrNoRows) {
		return Nucleo{}, ErrNucleoNaoEncontrado
	}
	return n, err
}

func (s *Postgres) BuscarNucleos(ctx context.Context, q string) ([]NucleoLista, error) {
	q = strings.TrimSpace(q)
	rows, err := s.pool.Query(ctx, `
		SELECT n.id::text, n.responsavel_legal, n.whatsapp, n.cesta,
		       n.logradouro, n.numero, n.complemento, n.bairro, n.cidade, n.uf, n.cep,
		       (SELECT count(*) FROM assistidos a WHERE a.nucleo_id = n.id)
		FROM nucleos_familiares n
		WHERE $1 = '' OR n.responsavel_legal ILIKE '%%'||$1||'%%'
		   OR n.whatsapp ILIKE '%%'||$1||'%%'
		   OR n.cidade ILIKE '%%'||$1||'%%'
		   OR n.logradouro ILIKE '%%'||$1||'%%'
		ORDER BY n.responsavel_legal
	`, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []NucleoLista{}
	for rows.Next() {
		var item NucleoLista
		if err := rows.Scan(&item.ID, &item.ResponsavelLegal, &item.WhatsApp, &item.Cesta,
			&item.Endereco.Logradouro, &item.Endereco.Numero, &item.Endereco.Complemento,
			&item.Endereco.Bairro, &item.Endereco.Cidade, &item.Endereco.UF, &item.Endereco.CEP,
			&item.Assistidos); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Postgres) LerNucleo(ctx context.Context, id string) (NucleoLista, error) {
	var item NucleoLista
	err := s.pool.QueryRow(ctx, `
		SELECT n.id::text, n.responsavel_legal, n.whatsapp, n.cesta,
		       n.logradouro, n.numero, n.complemento, n.bairro, n.cidade, n.uf, n.cep,
		       (SELECT count(*) FROM assistidos a WHERE a.nucleo_id = n.id)
		FROM nucleos_familiares n WHERE n.id = $1
	`, id).Scan(&item.ID, &item.ResponsavelLegal, &item.WhatsApp, &item.Cesta,
		&item.Endereco.Logradouro, &item.Endereco.Numero, &item.Endereco.Complemento,
		&item.Endereco.Bairro, &item.Endereco.Cidade, &item.Endereco.UF, &item.Endereco.CEP,
		&item.Assistidos)
	if errors.Is(err, pgx.ErrNoRows) {
		return NucleoLista{}, ErrNucleoNaoEncontrado
	}
	return item, err
}

func (s *Postgres) ListarUnidades(ctx context.Context, incluirInativos bool) ([]Unidade, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, nome, ativo FROM unidades
		WHERE ativo OR $1
		ORDER BY nome
	`, incluirInativos)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Unidade{}
	for rows.Next() {
		var u Unidade
		if err := rows.Scan(&u.ID, &u.Nome, &u.Ativo); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *Postgres) CriarUnidade(ctx context.Context, nome string) (Unidade, error) {
	u := Unidade{ID: MakeSlug(nome), Nome: nome, Ativo: true}
	_, err := s.pool.Exec(ctx, `INSERT INTO unidades (id, nome, ativo) VALUES ($1, $2, $3)`, u.ID, u.Nome, u.Ativo)
	if err != nil {
		return Unidade{}, err
	}
	return u, nil
}

func (s *Postgres) AtualizarUnidade(ctx context.Context, id, nome string) (Unidade, error) {
	u, err := s.GetUnidade(ctx, id)
	if err != nil {
		return Unidade{}, err
	}
	u.Nome = nome
	_, err = s.pool.Exec(ctx, `UPDATE unidades SET nome = $2 WHERE id = $1`, id, nome)
	if err != nil {
		return Unidade{}, err
	}
	return u, nil
}

func (s *Postgres) MudarAtivoUnidade(ctx context.Context, id string, ativo bool) (Unidade, error) {
	u, err := s.GetUnidade(ctx, id)
	if err != nil {
		return Unidade{}, err
	}
	u.Ativo = ativo
	_, err = s.pool.Exec(ctx, `UPDATE unidades SET ativo = $2 WHERE id = $1`, id, ativo)
	if err != nil {
		return Unidade{}, err
	}
	return u, nil
}

func (s *Postgres) GetUnidade(ctx context.Context, id string) (Unidade, error) {
	var u Unidade
	err := s.pool.QueryRow(ctx, `SELECT id, nome, ativo FROM unidades WHERE id = $1`, id).Scan(&u.ID, &u.Nome, &u.Ativo)
	if errors.Is(err, pgx.ErrNoRows) {
		return Unidade{}, ErrUnidadeInvalida
	}
	return u, err
}

func (s *Postgres) ListarOficinas(ctx context.Context, incluirInativos bool) ([]Oficina, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, nome, ativo FROM oficinas
		WHERE ativo OR $1
		ORDER BY nome
	`, incluirInativos)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Oficina{}
	for rows.Next() {
		var o Oficina
		if err := rows.Scan(&o.ID, &o.Nome, &o.Ativo); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (s *Postgres) CriarOficina(ctx context.Context, nome string) (Oficina, error) {
	o := Oficina{ID: MakeSlug(nome), Nome: nome, Ativo: true}
	_, err := s.pool.Exec(ctx, `INSERT INTO oficinas (id, nome, ativo) VALUES ($1, $2, $3)`, o.ID, o.Nome, o.Ativo)
	if err != nil {
		return Oficina{}, err
	}
	return o, nil
}

func (s *Postgres) AtualizarOficina(ctx context.Context, id, nome string) (Oficina, error) {
	o, err := s.GetOficina(ctx, id)
	if err != nil {
		return Oficina{}, err
	}
	o.Nome = nome
	_, err = s.pool.Exec(ctx, `UPDATE oficinas SET nome = $2 WHERE id = $1`, id, nome)
	if err != nil {
		return Oficina{}, err
	}
	return o, nil
}

func (s *Postgres) MudarAtivoOficina(ctx context.Context, id string, ativo bool) (Oficina, error) {
	o, err := s.GetOficina(ctx, id)
	if err != nil {
		return Oficina{}, err
	}
	o.Ativo = ativo
	_, err = s.pool.Exec(ctx, `UPDATE oficinas SET ativo = $2 WHERE id = $1`, id, ativo)
	if err != nil {
		return Oficina{}, err
	}
	return o, nil
}

func (s *Postgres) GetOficina(ctx context.Context, id string) (Oficina, error) {
	var o Oficina
	err := s.pool.QueryRow(ctx, `SELECT id, nome, ativo FROM oficinas WHERE id = $1`, id).Scan(&o.ID, &o.Nome, &o.Ativo)
	if errors.Is(err, pgx.ErrNoRows) {
		return Oficina{}, ErrOficinaInvalida
	}
	return o, err
}

func DatabaseURL() string {
	if u := os.Getenv("DATABASE_URL"); u != "" {
		return u
	}
	return "postgres://nacao:nacao@localhost:5432/nacao?sslmode=disable"
}
