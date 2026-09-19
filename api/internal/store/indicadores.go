package store

type SerieIndicador struct {
	ID          string `json:"id"`
	Nome        string `json:"nome"`
	Quantidade  int    `json:"quantidade"`
}

type Indicadores struct {
	UnidadeID  string            `json:"unidade_id"`
	Familias   int               `json:"familias"`
	Assistidos int               `json:"assistidos"`
	Cestas     int               `json:"cestas"`
	PorUnidade []SerieIndicador  `json:"por_unidade"`
	PorOficina []SerieIndicador  `json:"por_oficina"`
}

func CalcularIndicadores(cadastros []Cadastro, unidadeID string, unidades []Unidade, oficinas []Oficina) Indicadores {
	out := Indicadores{
		UnidadeID:  unidadeID,
		PorUnidade: make([]SerieIndicador, 0, len(unidades)),
		PorOficina: make([]SerieIndicador, 0, len(oficinas)),
	}
	filtrados := make([]Cadastro, 0)
	for _, c := range cadastros {
		if unidadeID != "" && c.UnidadeID != unidadeID {
			continue
		}
		filtrados = append(filtrados, c)
	}
	out.Assistidos = len(filtrados)
	nucleos := map[string]Nucleo{}
	for _, c := range filtrados {
		nucleos[c.NucleoID] = c.Nucleo
	}
	out.Familias = len(nucleos)
	for _, n := range nucleos {
		if n.Cesta {
			out.Cestas++
		}
	}
	contU := map[string]int{}
	contO := map[string]int{}
	for _, c := range filtrados {
		if c.UnidadeID != "" {
			contU[c.UnidadeID]++
		}
		if !c.Aluno {
			continue
		}
		for _, oid := range c.Oficinas {
			contO[oid]++
		}
	}
	for _, u := range unidades {
		out.PorUnidade = append(out.PorUnidade, SerieIndicador{ID: u.ID, Nome: u.Nome, Quantidade: contU[u.ID]})
	}
	for _, o := range oficinas {
		out.PorOficina = append(out.PorOficina, SerieIndicador{ID: o.ID, Nome: o.Nome, Quantidade: contO[o.ID]})
	}
	return out
}
