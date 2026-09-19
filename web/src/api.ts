export type Operador = {
  id: string;
  email: string;
};

export async function api<T>(path: string, init: RequestInit = {}): Promise<{ ok: boolean; status: number; data: T }> {
  const headers = new Headers(init.headers);
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const res = await fetch(path, { ...init, headers, credentials: "include" });
  const text = await res.text();
  const data = (text ? JSON.parse(text) : null) as T;
  return { ok: res.ok, status: res.status, data };
}

export function lerSessao() {
  return api<Operador>("/api/me");
}

export function entrar(email: string, senha: string) {
  return api<Operador>("/api/login", {
    method: "POST",
    body: JSON.stringify({ email, senha }),
  });
}

export function sair() {
  return api<null>("/api/logout", { method: "POST" });
}

export type Endereco = {
  logradouro: string;
  numero: string;
  complemento: string;
  bairro: string;
  cidade: string;
  uf: string;
  cep: string;
};

export type Nucleo = {
  id: string;
  responsavel_legal: string;
  whatsapp: string;
  cesta: boolean;
  endereco: Endereco;
};

export type Cadastro = {
  id: string;
  nucleo_id: string;
  nome: string;
  nome_exibicao: string;
  cpf: string;
  estrangeiro: boolean;
  pais_origem: string;
  unidade_id: string;
  oficinas: string[];
  aluno: boolean;
  nucleo: Nucleo;
  atendimento: {
    id: string;
    data: string;
    relato: string;
    itens_entregues: string;
  };
};

export type Catalogos = {
  unidades: { id: string; nome: string }[];
  oficinas: { id: string; nome: string }[];
};

export type Unidade = {
  id: string;
  nome: string;
  ativo: boolean;
};

export type Oficina = {
  id: string;
  nome: string;
  ativo: boolean;
};


export type ErroCampo = { campo?: string; erro?: string };

export type CadastroPedido = {
  nome: string;
  cpf: string;
  estrangeiro: boolean;
  pais_origem: string;
  unidade_id: string;
  oficinas: string[];
  nucleo: {
    id?: string;
    responsavel_legal: string;
    whatsapp: string;
    cesta: boolean;
    endereco: Endereco;
  };
  atendimento: { data: string; relato: string; itens_entregues: string };
};

export function criarCadastro(body: CadastroPedido) {
  return api<Cadastro & ErroCampo>("/api/assistidos", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function atualizarCadastro(id: string, body: CadastroPedido) {
  return api<Cadastro & ErroCampo>(`/api/assistidos/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function listarCadastros() {
  return api<{ itens: Cadastro[]; total: number; pagina: number; por_pagina: number }>("/api/assistidos");
}

export function consultarCadastros(params: { q?: string; unidade_id?: string; pagina?: number; por_pagina?: number }) {
  const qs = new URLSearchParams();
  if (params.q) qs.set("q", params.q);
  if (params.unidade_id) qs.set("unidade_id", params.unidade_id);
  if (params.pagina) qs.set("pagina", String(params.pagina));
  if (params.por_pagina) qs.set("por_pagina", String(params.por_pagina));
  const suffix = qs.toString() ? `?${qs}` : "";
  return api<{ itens: Cadastro[]; total: number; pagina: number; por_pagina: number }>(`/api/assistidos${suffix}`);
}

export function lerCadastro(id: string) {
  return api<Cadastro & ErroCampo>(`/api/assistidos/${id}`);
}

export type SerieIndicador = { id: string; nome: string; quantidade: number };

export type Indicadores = {
  unidade_id: string;
  familias: number;
  assistidos: number;
  cestas: number;
  por_unidade: SerieIndicador[];
  por_oficina: SerieIndicador[];
};

export function lerIndicadores(unidade_id?: string) {
  const qs = unidade_id ? `?unidade_id=${encodeURIComponent(unidade_id)}` : "";
  return api<Indicadores>(`/api/indicadores${qs}`);
}

export async function baixarPDF(path: string, filename: string) {
  const res = await fetch(path, { credentials: "include" });
  if (!res.ok) return false;
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
  return true;
}

export function lerCatalogos() {
  return api<Catalogos>("/api/catalogos");
}

export function listarUnidades() {
  return api<Unidade[]>("/api/unidades");
}

export function criarUnidade(nome: string) {
  return api<Unidade & ErroCampo>("/api/unidades", {
    method: "POST",
    body: JSON.stringify({ nome }),
  });
}

export function atualizarUnidade(id: string, nome: string) {
  return api<Unidade & ErroCampo>(`/api/unidades/${id}`, {
    method: "PUT",
    body: JSON.stringify({ nome }),
  });
}

export function alternarAtivoUnidade(id: string, ativo: boolean) {
  return api<Unidade & ErroCampo>(`/api/unidades/${id}/ativo`, {
    method: "PUT",
    body: JSON.stringify({ ativo }),
  });
}

export function listarOficinas() {
  return api<Oficina[]>("/api/oficinas");
}

export function criarOficina(nome: string) {
  return api<Oficina & ErroCampo>("/api/oficinas", {
    method: "POST",
    body: JSON.stringify({ nome }),
  });
}

export function atualizarOficina(id: string, nome: string) {
  return api<Oficina & ErroCampo>(`/api/oficinas/${id}`, {
    method: "PUT",
    body: JSON.stringify({ nome }),
  });
}

export function alternarAtivoOficina(id: string, ativo: boolean) {
  return api<Oficina & ErroCampo>(`/api/oficinas/${id}/ativo`, {
    method: "PUT",
    body: JSON.stringify({ ativo }),
  });
}

export type NucleoLista = Nucleo & { assistidos: number };

export function buscarNucleos(q: string) {
  const qs = q ? `?q=${encodeURIComponent(q)}` : "";
  return api<{ itens: NucleoLista[] }>(`/api/nucleos${qs}`);
}
