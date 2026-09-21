export function soDigitos(valor: string): string {
  return valor.replace(/\D/g, "");
}

export function mascararCPF(valor: string): string {
  const d = soDigitos(valor).slice(0, 11);
  const a = d.slice(0, 3);
  const b = d.slice(3, 6);
  const c = d.slice(6, 9);
  const e = d.slice(9, 11);
  if (d.length <= 3) return a;
  if (d.length <= 6) return `${a}.${b}`;
  if (d.length <= 9) return `${a}.${b}.${c}`;
  return `${a}.${b}.${c}-${e}`;
}

export function formatarCPF(valor: string): string {
  const d = soDigitos(valor);
  if (!d) return "";
  if (d.length !== 11) return valor.trim();
  return mascararCPF(d);
}

export function mascararCEP(valor: string): string {
  const d = soDigitos(valor).slice(0, 8);
  if (d.length <= 5) return d;
  return `${d.slice(0, 5)}-${d.slice(5)}`;
}

export function mascararData(valor: string): string {
  const d = soDigitos(valor).slice(0, 8);
  const dd = d.slice(0, 2);
  const mm = d.slice(2, 4);
  const aaaa = d.slice(4, 8);
  if (d.length <= 2) return dd;
  if (d.length <= 4) return `${dd}/${mm}`;
  return `${dd}/${mm}/${aaaa}`;
}

export function dataCalendarioValida(ano: number, mes: number, dia: number): boolean {
  if (ano < 1900 || ano > 2100 || mes < 1 || mes > 12 || dia < 1 || dia > 31) return false;
  const dt = new Date(ano, mes - 1, dia);
  return dt.getFullYear() === ano && dt.getMonth() === mes - 1 && dt.getDate() === dia;
}

/** Vazio → ""; inválida → null; válida → YYYY-MM-DD. */
export function dataBRParaISO(valor: string): string | null {
  const t = valor.trim();
  if (!t) return "";
  const d = soDigitos(t);
  if (d.length !== 8) return null;
  const dia = Number(d.slice(0, 2));
  const mes = Number(d.slice(2, 4));
  const ano = Number(d.slice(4, 8));
  if (!dataCalendarioValida(ano, mes, dia)) return null;
  return `${String(ano).padStart(4, "0")}-${String(mes).padStart(2, "0")}-${String(dia).padStart(2, "0")}`;
}

export function isoParaDataBR(iso: string): string {
  const t = iso.trim();
  const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(t);
  if (!m) return t;
  return `${m[3]}/${m[2]}/${m[1]}`;
}

export function hojeBR(): string {
  return isoParaDataBR(hojeISO());
}

export function hojeISO(): string {
  const d = new Date();
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

export function formatarDataBR(iso: string): string {
  const t = iso.trim();
  if (!t) return "";
  return isoParaDataBR(t);
}

export function idadeEmAnos(iso: string, hoje = hojeISO()): number | null {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(iso.trim());
  if (!m) return null;
  const ny = Number(m[1]);
  const nm = Number(m[2]);
  const nd = Number(m[3]);
  const h = /^(\d{4})-(\d{2})-(\d{2})$/.exec(hoje);
  if (!h) return null;
  const hy = Number(h[1]);
  const hm = Number(h[2]);
  const hd = Number(h[3]);
  let idade = hy - ny;
  if (hm < nm || (hm === nm && hd < nd)) idade--;
  return idade;
}

export function ehMaiorDeIdade(iso: string, hoje = hojeISO()): boolean {
  const idade = idadeEmAnos(iso, hoje);
  return idade !== null && idade >= 18;
}
