
package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Arthur-Bamberg/nacao-assistidos/api/internal/store"
)

const cookieSessao = "nacao_sessao"

type Server struct {
	store      store.Store
	secure     bool
	novaSessao func() (string, error)
	httpClient *http.Client
	viacepBase string
}

func New(st store.Store) *Server {
	return &Server{
		store:  st,
		secure: os.Getenv("COOKIE_SECURE") == "true",
		novaSessao: func() (string, error) {
			b := make([]byte, 32)
			if _, err := rand.Read(b); err != nil {
				return "", err
			}
			return hex.EncodeToString(b), nil
		},
		httpClient: &http.Client{Timeout: 4 * time.Second},
		viacepBase: "https://viacep.com.br",
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/saude", s.saude)
	mux.HandleFunc("POST /api/login", s.login)
	mux.HandleFunc("POST /api/logout", s.logout)
	mux.HandleFunc("GET /api/me", s.requerSessao(s.me))
	mux.HandleFunc("GET /api/catalogos", s.requerSessao(s.catalogos))
	mux.HandleFunc("POST /api/assistidos", s.requerSessao(s.criarCadastro))
	mux.HandleFunc("GET /api/assistidos", s.requerSessao(s.listarCadastros))
	mux.HandleFunc("GET /api/assistidos/pdf", s.requerSessao(s.consultaPDF))
	mux.HandleFunc("GET /api/assistidos/{id}", s.requerSessao(s.lerCadastro))
	mux.HandleFunc("GET /api/assistidos/{id}/atendimentos", s.requerSessao(s.listarAtendimentos))
	mux.HandleFunc("PUT /api/assistidos/{id}", s.requerSessao(s.atualizarCadastro))
	mux.HandleFunc("GET /api/cep/{cep}", s.requerSessao(s.consultarCEP))
	mux.HandleFunc("GET /api/nucleos", s.requerSessao(s.buscarNucleos))
	mux.HandleFunc("GET /api/nucleos/{id}", s.requerSessao(s.lerNucleo))
	mux.HandleFunc("GET /api/indicadores", s.requerSessao(s.indicadores))
	mux.HandleFunc("GET /api/indicadores/pdf", s.requerSessao(s.indicadoresPDF))
	mux.HandleFunc("GET /api/unidades", s.requerSessao(s.listarUnidades))
	mux.HandleFunc("POST /api/unidades", s.requerSessao(s.criarUnidade))
	mux.HandleFunc("PUT /api/unidades/{id}", s.requerSessao(s.atualizarUnidade))
	mux.HandleFunc("PUT /api/unidades/{id}/ativo", s.requerSessao(s.mudarAtivoUnidade))
	mux.HandleFunc("GET /api/oficinas", s.requerSessao(s.listarOficinas))
	mux.HandleFunc("POST /api/oficinas", s.requerSessao(s.criarOficina))
	mux.HandleFunc("PUT /api/oficinas/{id}", s.requerSessao(s.atualizarOficina))
	mux.HandleFunc("PUT /api/oficinas/{id}/ativo", s.requerSessao(s.mudarAtivoOficina))
	return mux
}

func (s *Server) saude(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type loginBody struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "pedido inválido"})
		return
	}
	email := strings.TrimSpace(strings.ToLower(body.Email))
	op, err := s.store.Autenticar(r.Context(), email, body.Senha)
	if errors.Is(err, store.ErrCredenciaisInvalidas) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"erro": "credenciais inválidas"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": "falha ao autenticar"})
		return
	}
	token, err := s.novaSessao()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": "falha ao autenticar"})
		return
	}
	if err := s.store.CriarSessao(r.Context(), op.ID, token, 12*time.Hour); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": "falha ao autenticar"})
		return
	}
	s.gravarCookie(w, token, 12*time.Hour)
	writeJSON(w, http.StatusOK, map[string]string{"id": op.ID, "email": op.Email})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(cookieSessao)
	if err == nil {
		_ = s.store.EncerrarSessao(r.Context(), c.Value)
	}
	s.gravarCookie(w, "", -time.Hour)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	op := operadorDoContexto(r.Context())
	writeJSON(w, http.StatusOK, map[string]string{"id": op.ID, "email": op.Email})
}

type ctxKey int

const operadorKey ctxKey = 1

func (s *Server) requerSessao(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(cookieSessao)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autenticado"})
			return
		}
		op, err := s.store.OperadorDaSessao(r.Context(), c.Value)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"erro": "não autenticado"})
			return
		}
		ctx := context.WithValue(r.Context(), operadorKey, op)
		next(w, r.WithContext(ctx))
	}
}

func operadorDoContexto(ctx context.Context) store.Operador {
	op, _ := ctx.Value(operadorKey).(store.Operador)
	return op
}

func (s *Server) gravarCookie(w http.ResponseWriter, token string, maxAge time.Duration) {
	c := &http.Cookie{
		Name:     cookieSessao,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.secure,
	}
	if maxAge < 0 {
		c.MaxAge = -1
		c.Expires = time.Unix(0, 0)
	} else {
		c.MaxAge = int(maxAge.Seconds())
	}
	http.SetCookie(w, c)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type criarUnidadeBody struct {
	Nome string `json:"nome"`
}

func (s *Server) listarUnidades(w http.ResponseWriter, r *http.Request) {
	unidades, err := s.store.ListarUnidades(r.Context(), true) // Include inactive for admin listing
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, unidades)
}

func (s *Server) criarUnidade(w http.ResponseWriter, r *http.Request) {
	var body criarUnidadeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "pedido inválido"})
		return
	}
	unidade, err := s.store.CriarUnidade(r.Context(), body.Nome)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, unidade)
}

type atualizarUnidadeBody struct {
	Nome string `json:"nome"`
}

func (s *Server) atualizarUnidade(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body atualizarUnidadeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "pedido inválido"})
		return
	}
	unidade, err := s.store.AtualizarUnidade(r.Context(), id, body.Nome)
	if errors.Is(err, store.ErrUnidadeInvalida) {
		writeJSON(w, http.StatusNotFound, map[string]string{"erro": "unidade não encontrada"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, unidade)
}

type mudarAtivoBody struct {
	Ativo bool `json:"ativo"`
}

func (s *Server) mudarAtivoUnidade(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body mudarAtivoBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "pedido inválido"})
		return
	}
	unidade, err := s.store.MudarAtivoUnidade(r.Context(), id, body.Ativo)
	if errors.Is(err, store.ErrUnidadeInvalida) {
		writeJSON(w, http.StatusNotFound, map[string]string{"erro": "unidade não encontrada"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, unidade)
}

type criarOficinaBody struct {
	Nome string `json:"nome"`
}

func (s *Server) listarOficinas(w http.ResponseWriter, r *http.Request) {
	oficinas, err := s.store.ListarOficinas(r.Context(), true) // Include inactive for admin listing
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, oficinas)
}

func (s *Server) criarOficina(w http.ResponseWriter, r *http.Request) {
	var body criarOficinaBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "pedido inválido"})
		return
	}
	oficina, err := s.store.CriarOficina(r.Context(), body.Nome)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, oficina)
}

type atualizarOficinaBody struct {
	Nome string `json:"nome"`
}

func (s *Server) atualizarOficina(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body atualizarOficinaBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "pedido inválido"})
		return
	}
	oficina, err := s.store.AtualizarOficina(r.Context(), id, body.Nome)
	if errors.Is(err, store.ErrOficinaInvalida) {
		writeJSON(w, http.StatusNotFound, map[string]string{"erro": "oficina não encontrada"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, oficina)
}

func (s *Server) mudarAtivoOficina(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body mudarAtivoBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "pedido inválido"})
		return
	}
	oficina, err := s.store.MudarAtivoOficina(r.Context(), id, body.Ativo)
	if errors.Is(err, store.ErrOficinaInvalida) {
		writeJSON(w, http.StatusNotFound, map[string]string{"erro": "oficina não encontrada"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, oficina)
}
