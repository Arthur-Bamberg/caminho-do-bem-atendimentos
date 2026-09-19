# Cadastro de Assistidos

PWA das filiais para acompanhar Assistidos, Núcleos Familiares e Atendimentos.

Glossário: [`CONTEXT.md`](CONTEXT.md). PRD: [`docs/PRD.md`](docs/PRD.md).

## Como rodar

Stack completa (Postgres, API e PWA):

```sh
docker compose up --build
```

PWA em `http://localhost:3000` (a API fica em `/api` no mesmo endereço).

Para desenvolver fora do Compose: `make postgres`, depois `make api` e `make web`. Sem Postgres, a API sobe com armazenamento em memória só para desenvolvimento.

Conta seed: `operador@nacao.local` / `nacao-dev`.

Testes da API: `make test-api`.
