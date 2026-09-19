CREATE TABLE IF NOT EXISTS operadores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    senha_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessoes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operador_id UUID NOT NULL REFERENCES operadores (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expira_em TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS sessoes_token_hash_idx ON sessoes (token_hash);
CREATE INDEX IF NOT EXISTS sessoes_expira_em_idx ON sessoes (expira_em);
