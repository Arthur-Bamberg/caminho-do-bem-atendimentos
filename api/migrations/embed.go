package migrations

import "embed"

//go:embed *.sql
var SQL embed.FS

func Operadores() string {
	return must("001_operadores.sql")
}

func Cadastro() string {
	return must("002_cadastro.sql")
}

func UnidadeOficinas() string {
	return must("003_unidade_oficinas.sql")
}

func Nucleo() string {
	return must("004_nucleo.sql")
}

func CatalogoCrud() string {
	return must("005_catalogo_crud.sql")
}

func must(name string) string {
	b, err := SQL.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return string(b)
}
