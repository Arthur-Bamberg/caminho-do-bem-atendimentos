import { useEffect, useState } from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { lerSessao, type Operador } from "./api";
import { CadastroPage } from "./CadastroPage";
import { ConsultaPage } from "./ConsultaPage";
import { Casca } from "./Casca";
import { LoginPage } from "./LoginPage";
import { PainelPage } from "./PainelPage";
import { UnidadesPage } from "./UnidadesPage";
import { OficinasPage } from "./OficinasPage";
import { RequerSessao } from "./RequerSessao";
import { SessaoContext } from "./sessao";

export default function App() {
  const [operador, setOperador] = useState<Operador | null>(null);
  const [pronto, setPronto] = useState(false);

  useEffect(() => {
    lerSessao()
      .then((res) => {
        if (res.ok) setOperador(res.data);
      })
      .finally(() => setPronto(true));
  }, []);

  if (!pronto) {
    return (
      <div className="login-page">
        <p>Carregando…</p>
      </div>
    );
  }

  return (
    <SessaoContext.Provider value={{ operador, setOperador }}>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={operador ? <Navigate to="/cadastro" replace /> : <LoginPage />} />
          <Route element={<RequerSessao />}>
            <Route element={<Casca />}>
              <Route path="/cadastro" element={<CadastroPage />} />
              <Route path="/consulta" element={<ConsultaPage />} />
              <Route path="/painel" element={<PainelPage />} />
              <Route path="/unidades" element={<UnidadesPage />} />
              <Route path="/oficinas" element={<OficinasPage />} />
            </Route>
          </Route>
          <Route path="*" element={<Navigate to={operador ? "/cadastro" : "/login"} replace />} />
        </Routes>
      </BrowserRouter>
    </SessaoContext.Provider>
  );
}
