export function PaginaEmConstrucao({ titulo }: { titulo: string }) {
  return (
    <section className="card placeholder-page">
      <h1>{titulo}</h1>
      <p>Esta tela entra nas próximas fatias. A casca já exige sessão de Operador.</p>
    </section>
  );
}
