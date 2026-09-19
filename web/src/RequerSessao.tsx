import { Navigate, Outlet } from "react-router-dom";
import { useSessao } from "./sessao";

export function RequerSessao() {
  const { operador } = useSessao();
  if (!operador) {
    return <Navigate to="/login" replace />;
  }
  return <Outlet />;
}
