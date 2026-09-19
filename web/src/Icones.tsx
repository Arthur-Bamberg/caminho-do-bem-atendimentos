import type { ReactNode } from "react";

function Icone({ children }: { children: ReactNode }) {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
      {children}
    </svg>
  );
}

export function IconeCadastro() {
  return (
    <Icone>
      <path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z" />
      <path d="M14 3v5h5" />
      <path d="M9 13h6M9 17h4" />
    </Icone>
  );
}

export function IconeConsulta() {
  return (
    <Icone>
      <circle cx="11" cy="11" r="6.5" />
      <path d="m16 16 4 4" />
    </Icone>
  );
}

export function IconePainel() {
  return (
    <Icone>
      <path d="M4 19V9M10 19V5M16 19v-7M22 19H2" />
    </Icone>
  );
}

export function IconeUnidades() {
  return (
    <Icone>
      <path d="M4 20V8l6-4 6 4v12" />
      <path d="M14 20v-8h6v8" />
      <path d="M9 20v-4h2v4" />
    </Icone>
  );
}

export function IconeOficinas() {
  return (
    <Icone>
      <circle cx="8" cy="8" r="2.5" />
      <circle cx="16" cy="8" r="2.5" />
      <path d="M4 18c.4-2.6 2.2-4 4-4s3.6 1.4 4 4" />
      <path d="M12 18c.4-2.6 2.2-4 4-4s3.6 1.4 4 4" />
    </Icone>
  );
}

export function IconeSair() {
  return (
    <Icone>
      <path d="M10 5H6a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h4" />
      <path d="M13 16l4-4-4-4M17 12H9" />
    </Icone>
  );
}

export function IconeEditar() {
  return (
    <Icone>
      <path d="M4 20h4L19 9l-4-4L4 16z" />
      <path d="m15 5 4 4" />
    </Icone>
  );
}
