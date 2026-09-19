CREATE TABLE IF NOT EXISTS nucleos_familiares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

CREATE TABLE IF NOT EXISTS assistidos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nucleo_id UUID NOT NULL REFERENCES nucleos_familiares (id),
    nome TEXT NOT NULL DEFAULT '',
    cpf TEXT NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS assistidos_cpf_unico
    ON assistidos (cpf)
    WHERE cpf <> '';

CREATE TABLE IF NOT EXISTS atendimentos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assistido_id UUID NOT NULL REFERENCES assistidos (id),
    data DATE NOT NULL,
    relato TEXT NOT NULL DEFAULT '',
    itens_entregues TEXT NOT NULL DEFAULT ''
);
