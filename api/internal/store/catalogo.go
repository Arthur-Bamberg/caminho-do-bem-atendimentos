package store

import (
	"errors"
	"strings"
)

type Unidade struct {
	ID    string `json:"id"`
	Nome  string `json:"nome"`
	Ativo bool   `json:"ativo"`
}

type Oficina struct {
	ID    string `json:"id"`
	Nome  string `json:"nome"`
	Ativo bool   `json:"ativo"`
}

type Catalogos struct {
	Unidades []Unidade `json:"unidades"`
	Oficinas []Oficina `json:"oficinas"`
}

var ErrUnidadeInativa = errors.New("unidade inativa")
var ErrOficinaInativa = errors.New("oficina inativa")

func ValidarUnidade(u Unidade, permitirInativo bool) error {
	if u.ID == "" {
		return nil
	}
	if !permitirInativo && !u.Ativo {
		return ErrUnidadeInativa
	}
	return nil
}

func ValidarOficinas(oficinas []Oficina, ids []string, permitirInativo bool) ([]string, error) {
	if ids == nil {
		return []string{}, nil
	}
	visto := map[string]bool{}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		ok := false
		for _, o := range oficinas {
			if o.ID == id {
				ok = true
				if !permitirInativo && !o.Ativo {
					return nil, ErrOficinaInativa
				}
				break
			}
		}
		if !ok {
			return nil, ErrOficinaInvalida
		}
		if visto[id] {
			continue
		}
		visto[id] = true
		out = append(out, id)
	}
	return out, nil
}

func EhAluno(oficinas []string) bool {
	return len(oficinas) > 0
}

func NucleoPayloadVazio(n NucleoNovo) bool {
	return n.ResponsavelLegal == "" && n.WhatsApp == "" && !n.Cesta &&
		n.Endereco.Logradouro == "" && n.Endereco.Numero == "" && n.Endereco.Complemento == "" &&
		n.Endereco.Bairro == "" && n.Endereco.Cidade == "" && n.Endereco.UF == "" && n.Endereco.CEP == ""
}

func MakeSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "á", "a")
	slug = strings.ReplaceAll(slug, "ã", "a")
	slug = strings.ReplaceAll(slug, "à", "a")
	slug = strings.ReplaceAll(slug, "é", "e")
	slug = strings.ReplaceAll(slug, "ê", "e")
	slug = strings.ReplaceAll(slug, "í", "i")
	slug = strings.ReplaceAll(slug, "ó", "o")
	slug = strings.ReplaceAll(slug, "ô", "o")
	slug = strings.ReplaceAll(slug, "õ", "o")
	slug = strings.ReplaceAll(slug, "ú", "u")
	slug = strings.ReplaceAll(slug, "ç", "c")
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, ",", "")
	slug = strings.ReplaceAll(slug, "(", "")
	slug = strings.ReplaceAll(slug, ")", "")
	return slug
}
