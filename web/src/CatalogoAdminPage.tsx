import { useEffect, useState, type FormEvent, type ReactNode } from "react";
import type { ErroCampo } from "./api";
import { IconeEditar } from "./Icones";

export type ItemCatalogo = {
  id: string;
  nome: string;
  ativo: boolean;
};

type Resposta<T> = Promise<{ ok: boolean; data: T & Partial<ErroCampo> }>;

type Props = {
  titulo: string;
  icone: ReactNode;
  lede: string;
  rotuloNome: string;
  placeholder: string;
  tituloLista: string;
  vazio: string;
  rotuloCriar: string;
  listar: () => Promise<{ ok: boolean; data: ItemCatalogo[] | null }>;
  criar: (nome: string) => Resposta<ItemCatalogo>;
  atualizar: (id: string, nome: string) => Resposta<ItemCatalogo>;
  alternarAtivo: (id: string, ativo: boolean) => Resposta<ItemCatalogo>;
};

export function CatalogoAdminPage({
  titulo,
  icone,
  lede,
  rotuloNome,
  placeholder,
  tituloLista,
  vazio,
  rotuloCriar,
  listar,
  criar,
  atualizar,
  alternarAtivo,
}: Props) {
  const [itens, setItens] = useState<ItemCatalogo[]>([]);
  const [nome, setNome] = useState("");
  const [editandoId, setEditandoId] = useState<string | null>(null);
  const [erro, setErro] = useState<string>("");
  const [enviando, setEnviando] = useState(false);

  async function refresh() {
    const res = await listar();
    if (res.ok && Array.isArray(res.data)) {
      setItens(res.data);
      setErro("");
      return;
    }
    setErro("Não foi possível carregar a lista.");
  }

  useEffect(() => {
    void refresh();
  }, []);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    if (enviando) return;
    setEnviando(true);
    setErro("");
    const res = editandoId ? await atualizar(editandoId, nome) : await criar(nome);
    setEnviando(false);
    if (!res.ok) {
      setErro(res.data && "erro" in res.data && res.data.erro ? String(res.data.erro) : "Não foi possível salvar.");
      return;
    }
    setNome("");
    setEditandoId(null);
    await refresh();
  }

  async function handleAlternarAtivo(item: ItemCatalogo) {
    await alternarAtivo(item.id, !item.ativo);
    await refresh();
  }

  function handleEditar(item: ItemCatalogo) {
    setNome(item.nome);
    setEditandoId(item.id);
    setErro("");
  }

  return (
    <section className="card">
      <div className="titulo-com-icone">
        {icone}
        <h1>{titulo}</h1>
      </div>
      <p className="lede">{lede}</p>

      <form onSubmit={handleSubmit} noValidate className="catalogo-form">
        <label className="field">
          <span>{rotuloNome}</span>
          <input value={nome} onChange={(e) => setNome(e.target.value)} placeholder={placeholder} />
        </label>
        {erro ? (
          <p className="erro" role="alert">
            {erro}
          </p>
        ) : null}
        <div className="acoes">
          <button className="primary" type="submit" disabled={enviando}>
            {enviando ? "Salvando…" : editandoId ? "Salvar edição" : rotuloCriar}
          </button>
          {editandoId ? (
            <button
              className="ghost"
              type="button"
              onClick={() => {
                setEditandoId(null);
                setNome("");
                setErro("");
              }}
            >
              Cancelar
            </button>
          ) : null}
        </div>
      </form>

      <h2 className="lista-titulo">{tituloLista}</h2>
      {itens.length === 0 ? (
        <p className="lede">{vazio}</p>
      ) : (
        <div className="tabela-wrap">
          <table className="tabela">
            <thead>
              <tr>
                <th>Nome</th>
                <th>Situação</th>
                <th>Ações</th>
              </tr>
            </thead>
            <tbody>
              {itens.map((item) => (
                <tr key={item.id} className={item.ativo ? undefined : "linha-inativa"}>
                  <td>{item.nome}</td>
                  <td>{item.ativo ? "Ativo" : "Inativo"}</td>
                  <td>
                    <div className="acoes">
                      <button className="ghost" type="button" onClick={() => handleEditar(item)}>
                        <IconeEditar />
                        Editar
                      </button>
                      <button className="ghost" type="button" onClick={() => void handleAlternarAtivo(item)}>
                        {item.ativo ? "Inativar" : "Ativar"}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
