import { useEffect, useState } from "react";
import { baixarPDF, lerCatalogos, lerIndicadores, type Catalogos, type Indicadores, type SerieIndicador } from "./api";

const vazio: Indicadores = {
  unidade_id: "",
  familias: 0,
  assistidos: 0,
  cestas: 0,
  por_unidade: [],
  por_oficina: [],
};

function Grafico({ titulo, series }: { titulo: string; series: SerieIndicador[] }) {
  const max = Math.max(0, ...series.map((s) => s.quantidade));
  return (
    <section className="grafico" aria-labelledby={titulo.replace(/\s/g, "-")}>
      <h2 id={titulo.replace(/\s/g, "-")}>{titulo}</h2>
      {series.length === 0 ? (
        <p className="ok">Sem série neste filtro.</p>
      ) : (
        <ul className="barras">
          {series.map((s) => (
            <li key={s.id}>
              <div className="barra-rotulo">
                <span>{s.nome}</span>
                <span>{s.quantidade}</span>
              </div>
              <div className="barra-trilha" aria-hidden="true">
                <div className="barra-preenchimento" style={{ width: max ? `${(s.quantidade / max) * 100}%` : "0%" }} />
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

export function PainelPage() {
  const [unidadeId, setUnidadeId] = useState("");
  const [catalogos, setCatalogos] = useState<Catalogos>({ unidades: [], oficinas: [] });
  const [ind, setInd] = useState<Indicadores>(vazio);
  const [baixando, setBaixando] = useState(false);

  useEffect(() => {
    lerCatalogos().then((res) => {
      if (res.ok) setCatalogos(res.data);
    });
  }, []);

  useEffect(() => {
    lerIndicadores(unidadeId).then((res) => {
      if (res.ok) setInd(res.data);
    });
  }, [unidadeId]);

  async function onPDF() {
    setBaixando(true);
    const qs = unidadeId ? `?unidade_id=${encodeURIComponent(unidadeId)}` : "";
    await baixarPDF(`/api/indicadores/pdf${qs}`, "painel.pdf");
    setBaixando(false);
  }

  return (
    <section className="card painel">
      <h1>Painel de indicadores</h1>
      <p className="lede">Contagens da API. Famílias são Núcleos; Cesta conta uma vez por Núcleo. Aluno é Assistido com Oficina vigente.</p>
      <div className="consulta-filtros">
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
        <button className="ghost" type="button" disabled={baixando} onClick={() => void onPDF()}>
          {baixando ? "Gerando PDF…" : "Baixar PDF do painel"}
        </button>
      </div>
      <div className="kpis">
        <article className="kpi">
          <p>Famílias impactadas</p>
          <p className="kpi-num">{ind.familias}</p>
        </article>
        <article className="kpi">
          <p>Assistidos totais</p>
          <p className="kpi-num">{ind.assistidos}</p>
        </article>
        <article className="kpi">
          <p>Cestas básicas</p>
          <p className="kpi-num">{ind.cestas}</p>
        </article>
      </div>
      <Grafico titulo="Assistidos por Unidade" series={ind.por_unidade} />
      <Grafico titulo="Alunos por Oficina" series={ind.por_oficina} />
    </section>
  );
}
