import { useEffect, useRef, useState, type FormEvent } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import {
  atualizarCadastro,
  buscarNucleos,
  criarCadastro,
  lerCadastro,
  lerCatalogos,
  listarCadastros,
  type Cadastro,
  type Catalogos,
  type NucleoLista,
} from "./api";

function enderecoVazio() {
  return { logradouro: "", numero: "", complemento: "", bairro: "", cidade: "", uf: "", cep: "" };
}

function hojeISO() {
  const d = new Date();
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

export function CadastroPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const idUrl = params.get("id");
  const [nome, setNome] = useState("");
  const [cpf, setCpf] = useState("");
  const [estrangeiro, setEstrangeiro] = useState(false);
  const [paisOrigem, setPaisOrigem] = useState("");
  const [unidadeId, setUnidadeId] = useState("");
  const [oficinas, setOficinas] = useState<string[]>([]);
  const [responsavel, setResponsavel] = useState("");
  const [whatsapp, setWhatsapp] = useState("");
  const [cesta, setCesta] = useState(false);
  const [endereco, setEndereco] = useState(enderecoVazio);
  const [data, setData] = useState(hojeISO);
  const [relato, setRelato] = useState("");
  const [itens, setItens] = useState("");
  const [enviando, setEnviando] = useState(false);
  const [erroCpf, setErroCpf] = useState("");
  const [resumo, setResumo] = useState("");
  const [salvo, setSalvo] = useState<Cadastro | null>(null);
  const [editandoId, setEditandoId] = useState<string | null>(null);
  const [lista, setLista] = useState<Cadastro[]>([]);
  const [catalogos, setCatalogos] = useState<Catalogos>({ unidades: [], oficinas: [] });
  const [buscaNucleo, setBuscaNucleo] = useState("");
  const [nucleosEncontrados, setNucleosEncontrados] = useState<NucleoLista[]>([]);
  const [nucleoId, setNucleoId] = useState("");
  const resumoRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    listarCadastros().then((res) => {
      if (res.ok) setLista(res.data.itens ?? []);
    });
    lerCatalogos().then((res) => {
      if (res.ok) setCatalogos(res.data);
    });
  }, []);

  useEffect(() => {
    if (!idUrl) return;
    lerCadastro(idUrl).then((res) => {
      if (!res.ok) return;
      const c = res.data;
      setEditandoId(c.id);
      setNome(c.nome);
      setCpf(c.cpf);
      setEstrangeiro(c.estrangeiro);
      setPaisOrigem(c.pais_origem);
      setUnidadeId(c.unidade_id);
      setOficinas(c.oficinas ?? []);
      setNucleoId(c.nucleo.id);
      setResponsavel(c.nucleo.responsavel_legal);
      setWhatsapp(c.nucleo.whatsapp);
      setCesta(c.nucleo.cesta);
      setEndereco({ ...enderecoVazio(), ...c.nucleo.endereco });
      setData(hojeISO());
      setRelato("");
      setItens("");
      setSalvo(c);
    });
  }, [idUrl]);

  function limparFicha() {
    setEditandoId(null);
    setNome("");
    setCpf("");
    setEstrangeiro(false);
    setPaisOrigem("");
    setUnidadeId("");
    setOficinas([]);
    setResponsavel("");
    setWhatsapp("");
    setCesta(false);
    setEndereco(enderecoVazio());
    setData(hojeISO());
    setRelato("");
    setItens("");
    setErroCpf("");
    setResumo("");
    setNucleoId("");
    setBuscaNucleo("");
    setNucleosEncontrados([]);
    navigate("/cadastro");
  }

  function toggleOficina(id: string) {
    setOficinas((atual) => (atual.includes(id) ? atual.filter((x) => x !== id) : [...atual, id]));
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (enviando) return;
    setEnviando(true);
    setErroCpf("");
    setResumo("");
    const body = {
      nome,
      cpf,
      estrangeiro,
      pais_origem: estrangeiro ? paisOrigem : "",
      unidade_id: unidadeId,
      oficinas,
      nucleo: {
        id: nucleoId || undefined,
        responsavel_legal: responsavel,
        whatsapp,
        cesta,
        endereco,
      },
      atendimento: { data, relato, itens_entregues: itens },
    };
    const res = editandoId ? await atualizarCadastro(editandoId, body) : await criarCadastro(body);
    setEnviando(false);
    if (!res.ok) {
      const msg = res.data && "erro" in res.data ? String(res.data.erro) : "Não foi possível salvar.";
      if (res.data && "campo" in res.data && res.data.campo === "cpf") {
        setErroCpf(msg);
      }
      setResumo(msg);
      queueMicrotask(() => resumoRef.current?.focus());
      return;
    }
    setSalvo(res.data);
    // Limpa campos específicos após salvar
    setNome("");
    setCpf("");
    setEstrangeiro(false);
    setPaisOrigem("");
    setNucleoId("");
    setResponsavel("");
    setWhatsapp("");
    setCesta(false);
    setEndereco(enderecoVazio());
    setRelato("");
    setEditandoId(null);
    setResumo("");
    setErroCpf("");
    setBuscaNucleo("");
    setNucleosEncontrados([]);
    navigate("/cadastro");

    setLista((atual) => {
      const sem = atual.filter((item) => item.id !== res.data.id);
      return [...sem, res.data];
    });
  }

  return (
    <section className="card">
      <h1>Cadastro</h1>
      <p className="lede">Todos os campos são opcionais. Responsável, endereço, WhatsApp e Cesta pertencem ao Núcleo; Unidade e Oficinas, ao Assistido.</p>

      {resumo ? (
        <div className="erro" id="resumo-erros" tabIndex={-1} ref={resumoRef} role="alert">
          <a href="#cpf">{resumo}</a>
        </div>
      ) : null}

      {salvo ? (
        <p className="ok" role="status">
          Salvo: <strong>{salvo.nome_exibicao}</strong>
          {salvo.aluno ? " · Aluno" : ""}
          {salvo.nucleo.cesta ? " · Cesta" : ""}
        </p>
      ) : null}

      <form onSubmit={onSubmit} noValidate>
        <label className="field">
          <span>Unidade</span>
          <select name="unidade" value={unidadeId} onChange={(ev) => setUnidadeId(ev.target.value)}>
            <option value="">Sem Unidade</option>
            {catalogos.unidades.map((u) => (
              <option key={u.id} value={u.id}>
                {u.nome}
              </option>
            ))}
          </select>
        </label>
        <fieldset className="field oficinas">
          <legend>Oficinas vigentes</legend>
          {catalogos.oficinas.map((o) => (
            <label key={o.id} className="check">
              <input type="checkbox" checked={oficinas.includes(o.id)} onChange={() => toggleOficina(o.id)} />
              <span>{o.nome}</span>
            </label>
          ))}
        </fieldset>
        <label className="field">
          <span>Data do Atendimento</span>
          <input name="data" type="date" value={data} onChange={(ev) => setData(ev.target.value)} />
        </label>
        <label className="field">
          <span>Itens Entregues</span>
          <textarea name="itens" rows={3} value={itens} onChange={(ev) => setItens(ev.target.value)} placeholder="o que foi entregue neste Atendimento" />
        </label>
        <label className="field">
          <span>Nome</span>
          <input name="nome" value={nome} onChange={(ev) => setNome(ev.target.value)} placeholder="como a pessoa se apresenta" />
        </label>
        <label className="field">
          <span>CPF</span>
          <input
            id="cpf"
            name="cpf"
            inputMode="numeric"
            autoComplete="off"
            value={cpf}
            onChange={(ev) => setCpf(ev.target.value)}
            placeholder="somente se houver"
            aria-invalid={erroCpf ? true : undefined}
            aria-describedby={erroCpf ? "cpf-erro" : undefined}
          />
          {erroCpf ? (
            <p className="erro" id="cpf-erro">
              {erroCpf}
            </p>
          ) : null}
        </label>
        <label className="check">
          <input
            type="checkbox"
            checked={estrangeiro}
            onChange={(ev) => {
              setEstrangeiro(ev.target.checked);
              if (!ev.target.checked) setPaisOrigem("");
            }}
          />
          <span>Estrangeiro</span>
        </label>
        {estrangeiro ? (
          <label className="field">
            <span>País de Origem</span>
            <input
              name="pais_origem"
              value={paisOrigem}
              onChange={(ev) => setPaisOrigem(ev.target.value)}
              placeholder="somente se estrangeiro"
            />
          </label>
        ) : null}
        <label className="field">
          <span>Buscar Núcleo existente</span>
          <input
            value={buscaNucleo}
            onChange={(ev) => setBuscaNucleo(ev.target.value)}
            placeholder="responsável, WhatsApp ou cidade"
          />
        </label>
        <div className="acoes" style={{ marginBottom: 16 }}>
          <button
            className="ghost"
            type="button"
            onClick={async () => {
              const res = await buscarNucleos(buscaNucleo);
              if (res.ok) setNucleosEncontrados(res.data.itens ?? []);
            }}
          >
            Buscar Núcleo
          </button>
          {nucleoId ? (
            <button
              className="ghost"
              type="button"
              onClick={() => {
                setNucleoId("");
                setNucleosEncontrados([]);
              }}
            >
              Desvincular Núcleo
            </button>
          ) : null}
        </div>
        {nucleoId ? <p className="ok">Vinculado ao Núcleo {nucleoId.slice(0, 8)}…</p> : null}
        {nucleosEncontrados.length > 0 ? (
          <ul className="lista-minima">
            {nucleosEncontrados.map((n) => (
              <li key={n.id}>
                <button
                  className="ghost"
                  type="button"
                  onClick={() => {
                    setNucleoId(n.id);
                    setResponsavel(n.responsavel_legal);
                    setWhatsapp(n.whatsapp);
                    setCesta(n.cesta);
                    setEndereco({ ...enderecoVazio(), ...n.endereco });
                  }}
                >
                  Escolher: {n.responsavel_legal || "Sem responsável"} ({n.assistidos} assistidos)
                </button>
              </li>
            ))}
          </ul>
        ) : null}
        <label className="field">
          <span>Responsável Legal</span>
          <input name="responsavel" value={responsavel} onChange={(ev) => setResponsavel(ev.target.value)} placeholder="nome de exibição, não cria Assistido" />
        </label>
        <label className="field">
          <span>WhatsApp</span>
          <input name="whatsapp" inputMode="tel" value={whatsapp} onChange={(ev) => setWhatsapp(ev.target.value)} placeholder="do Núcleo" />
        </label>
        <label className="check nucleo-cesta">
          <input type="checkbox" checked={cesta} onChange={(ev) => setCesta(ev.target.checked)} />
          <span>Cesta Básica (está na lista agora)</span>
        </label>
        <fieldset className="field oficinas">
          <legend>Endereço do Núcleo</legend>
          <label className="field">
            <span>Logradouro</span>
            <input value={endereco.logradouro} onChange={(ev) => setEndereco({ ...endereco, logradouro: ev.target.value })} />
          </label>
          <div className="grade">
            <label className="field">
              <span>Número</span>
              <input value={endereco.numero} onChange={(ev) => setEndereco({ ...endereco, numero: ev.target.value })} />
            </label>
            <label className="field">
              <span>Complemento</span>
              <input value={endereco.complemento} onChange={(ev) => setEndereco({ ...endereco, complemento: ev.target.value })} />
            </label>
          </div>
          <label className="field">
            <span>Bairro</span>
            <input value={endereco.bairro} onChange={(ev) => setEndereco({ ...endereco, bairro: ev.target.value })} />
          </label>
          <div className="grade">
            <label className="field">
              <span>Cidade</span>
              <input value={endereco.cidade} onChange={(ev) => setEndereco({ ...endereco, cidade: ev.target.value })} />
            </label>
            <label className="field">
              <span>UF</span>
              <input value={endereco.uf} onChange={(ev) => setEndereco({ ...endereco, uf: ev.target.value })} maxLength={2} />
            </label>
            <label className="field">
              <span>CEP</span>
              <input value={endereco.cep} onChange={(ev) => setEndereco({ ...endereco, cep: ev.target.value })} />
            </label>
          </div>
        </fieldset>
        <label className="field">
          <span>Relato da Situação</span>
          <textarea name="relato" rows={4} value={relato} onChange={(ev) => setRelato(ev.target.value)} placeholder="o que aconteceu hoje" />
        </label>
        <div className="acoes">
          <button className="primary" type="submit" disabled={enviando}>
            {enviando ? "Salvando…" : "Salvar"}
          </button>
          <button className="ghost" type="button" onClick={limparFicha}>
            Novo Cadastro
          </button>
        </div>
      </form>

      <h2 className="lista-titulo">Cadastros desta sessão</h2>
      {lista.length === 0 ? (
        <p className="lede">Nenhum Cadastro ainda.</p>
      ) : (
        <ul className="lista-minima">
          {lista.map((item) => (
            <li key={item.id}>
              {item.nome_exibicao}
              {item.aluno ? " · Aluno" : ""}
              {item.nucleo?.cesta ? " · Cesta" : ""}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
