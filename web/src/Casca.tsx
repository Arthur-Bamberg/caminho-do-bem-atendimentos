import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { sair } from "./api";
import { IconeCadastro, IconeConsulta, IconeOficinas, IconePainel, IconeSair, IconeUnidades } from "./Icones";
import { Logo } from "./Logo";
import { useSessao } from "./sessao";

export function Casca() {
  const { setOperador } = useSessao();
  const navigate = useNavigate();

  async function onSair() {
    await sair();
    setOperador(null);
    navigate("/login", { replace: true });
  }

  return (
    <div className="app-shell">
      <header className="chrome">
        <span className="brand">
          <Logo />
          <strong>Cadastro de Assistidos</strong>
        </span>
        <nav className="nav" aria-label="Principal">
          <NavLink to="/cadastro">
            <IconeCadastro />
            Cadastro
          </NavLink>
          <NavLink to="/consulta">
            <IconeConsulta />
            Consulta
          </NavLink>
          <NavLink to="/painel">
            <IconePainel />
            Painel
          </NavLink>
          <NavLink to="/unidades">
            <IconeUnidades />
            Unidades
          </NavLink>
          <NavLink to="/oficinas">
            <IconeOficinas />
            Oficinas
          </NavLink>
          <button className="ghost" type="button" onClick={onSair}>
            <IconeSair />
            Sair
          </button>
        </nav>
      </header>
      <main className="main">
        <Outlet />
      </main>
    </div>
  );
}
