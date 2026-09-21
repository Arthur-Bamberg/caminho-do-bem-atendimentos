import { useEffect, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { consultarCadastros, baixarPDF, lerCatalogos, type Cadastro, type Catalogos } from "./api";
import { formatarCPF, formatarDataBR } from "./formatacao";

function nomeUnidade(id: string, cat: Catalogos) {
  return cat.unidades.find((u) => u.id === id)?.nome ?? (id || "—");
}

function nomesOficinas(ids: string[], cat: Catalogos) {
  if (!ids?.length) return "—";
  return ids.map((id) => cat.oficinas.find((o) => o.id === id)?.nome ?? id).join(", ");
}

function enderecoTexto(c: Cadastro) {
  const e = c.nucleo?.endereco;
  if (!e) return "—";
  const partes = [e.logradouro, e.numero, e.bairro, e.cidade, e.uf].filter(Boolean);
  return partes.length ? partes.join(", ") : "—";
}

export function ConsultaPage() {
  const [q, setQ] = useState("");
  const [unidadeId, setUnidadeId] = useState("");
  const [pagina, setPagina] = useState(1);
  const [catalogos, setCatalogos] = useState<Catalogos>({ unidades: [], oficinas: [] });
  const [itens, setItens] = useState<Cadastro[]>([]);
  const [total, setTotal] = useState(0);
  const [porPagina, setPorPagina] = useState(10);
  const [buscando, setBuscando] = useState(false);
  const [jaBuscou, setJaBuscou] = useState(false);
  const [baixando, setBaixando] = useState(false);

  useEffect(() => {
    lerCatalogos().then((res) => {
      if (res.ok) setCatalogos(res.data);
    });
  }, []);

  async function buscar(p = pagina) {
    setBuscando(true);
    const res = await consultarCadastros({ q, unidade_id: unidadeId, pagina: p, por_pagina: 10 });
    setBuscando(false);
    setJaBuscou(true);
    if (!res.ok) return;
    setItens(res.data.itens ?? []);
    setTotal(res.data.total);
    setPagina(res.data.pagina);
    setPorPagina(res.data.por_pagina);
  }

  useEffect(() => {
    void buscar(1);
  }, []);

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    void buscar(1);
  }

  const paginas = Math.max(1, Math.ceil(total / porPagina) || 1);

  return (
    <section className="card consulta">
      <h1>Consulta</h1>
      <p className="lede">Busque por nome, CPF ou responsável. Filtre por Unidade. O PDF traz as linhas desta página da busca.</p>
      <form className="consulta-filtros" onSubmit={onSubmit}>
        <label className="field">
          <span>Pesquisar</span>
          <input value={q} onChange={(ev) => setQ(ev.target.value)} placeholder="nome, CPF ou responsável" />
        </label>
        <label className="field">
          <span>Unidade</span>
          <select value={unidadeId} onChange={(ev) => setUnidadeId(ev.target.value)}>
            <option value="">Todas as Unidades</option>
            {catalogos.unidades.map((u) => (
              <option key={u.id} value={u.id}>
                {u.nome}
              </option>
            ))}
          </select>
        </label>
        <button className="primary" type="submit" disabled={buscando}>
          {buscando ? "Pesquisando…" : "Pesquisar"}
        </button>
        <button
          className="ghost"
          type="button"
          disabled={baixando}
          onClick={() => {
            const qs = new URLSearchParams();
            if (q) qs.set("q", q);
            if (unidadeId) qs.set("unidade_id", unidadeId);
            qs.set("pagina", String(pagina));
            qs.set("por_pagina", String(porPagina));
            setBaixando(true);
            void baixarPDF(`/api/assistidos/pdf?${qs}`, "consulta.pdf").finally(() => setBaixando(false));
          }}
        >
          {baixando ? "Gerando PDF…" : "Baixar PDF desta página"}
        </button>
      </form>

      {jaBuscou && total === 0 ? (
        <p className="ok" role="status">
          Nenhum Cadastro encontrado. Crie um novo em Cadastro.
        </p>
      ) : null}

      {itens.length > 0 ? (
        <div className="tabela-wrap">
          <table className="tabela">
            <thead>
              <tr>
                <th>Nome</th>
                <th>Filial</th>
                <th>Oficinas</th>
                <th>CPF</th>
                <th>Endereço</th>
                <th>Telefone</th>
                <th>Data</th>
                <th>Observações</th>
                <th>Ações</th>
              </tr>
            </thead>
            <tbody>
              {itens.map((c) => (
                <tr key={c.id}>
                  <td>{c.nome_exibicao}</td>
                  <td>{nomeUnidade(c.unidade_id, catalogos)}</td>
                  <td>{nomesOficinas(c.oficinas, catalogos)}</td>
                  <td>{formatarCPF(c.cpf) || "—"}</td>
                  <td>{enderecoTexto(c)}</td>
                  <td>{c.nucleo?.whatsapp || "—"}</td>
                  <td>{formatarDataBR(c.atendimento?.data || "") || "—"}</td>
                  <td>{c.atendimento?.relato || "—"}</td>
                  <td>
                    <Link className="ghost" to={`/cadastro?id=${c.id}`}>
                      Abrir ficha
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}

      {total > porPagina ? (
        <div className="acoes">
          <button className="ghost" type="button" disabled={pagina <= 1} onClick={() => void buscar(pagina - 1)}>
            Anterior
          </button>
          <span>
            Página {pagina} de {paginas}
          </span>
          <button className="ghost" type="button" disabled={pagina >= paginas} onClick={() => void buscar(pagina + 1)}>
            Próxima
          </button>
        </div>
      ) : null}
    </section>
  );
}
