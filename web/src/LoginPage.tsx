import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { entrar } from "./api";
import { Logo } from "./Logo";
import { useSessao } from "./sessao";

export function LoginPage() {
  const { setOperador } = useSessao();
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [senha, setSenha] = useState("");
  const [erro, setErro] = useState("");
  const [enviando, setEnviando] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (enviando) return;
    setEnviando(true);
    setErro("");
    const res = await entrar(email, senha);
    setEnviando(false);
    if (!res.ok) {
      setErro("E-mail ou senha inválidos.");
      return;
    }
    setOperador(res.data);
    navigate("/cadastro", { replace: true });
  }

  return (
    <div className="login-page">
      <form className="card" onSubmit={onSubmit} noValidate>
        <div className="brand" style={{ marginBottom: 16 }}>
          <Logo />
          <strong>Cadastro de Assistidos</strong>
        </div>
        <h1>Entrar</h1>
        <p className="lede">Somente Operadores provisionados acessam fichas e o painel.</p>
        {erro ? (
          <p className="erro" role="alert">
            {erro}
          </p>
        ) : null}
        <label className="field">
          <span>E-mail</span>
          <input
            id="email"
            name="email"
            type="email"
            autoComplete="username"
            value={email}
            onChange={(ev) => setEmail(ev.target.value)}
            placeholder="operador@nacao.local"
          />
        </label>
        <label className="field">
          <span>Senha</span>
          <input
            id="senha"
            name="password"
            type="password"
            autoComplete="current-password"
            value={senha}
            onChange={(ev) => setSenha(ev.target.value)}
          />
        </label>
        <button className="primary" type="submit" disabled={enviando}>
          {enviando ? "Entrando…" : "Entrar"}
        </button>
      </form>
    </div>
  );
}
