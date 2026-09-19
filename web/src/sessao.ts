import { createContext, useContext } from "react";
import type { Operador } from "./api";

export const SessaoContext = createContext<{
  operador: Operador | null;
  setOperador: (op: Operador | null) => void;
}>({ operador: null, setOperador: () => {} });

export function useSessao() {
  return useContext(SessaoContext);
}
