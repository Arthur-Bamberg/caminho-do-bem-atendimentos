CREATE TABLE IF NOT EXISTS unidades (
    id TEXT PRIMARY KEY,
    nome TEXT NOT NULL
);

INSERT INTO unidades (id, nome) VALUES
    ('centro', 'Centro'),
    ('norte', 'Norte'),
    ('sul', 'Sul')
ON CONFLICT (id) DO NOTHING;

ALTER TABLE assistidos ADD COLUMN IF NOT EXISTS unidade_id TEXT REFERENCES unidades (id);

CREATE TABLE IF NOT EXISTS assistido_oficinas (
    assistido_id UUID NOT NULL REFERENCES assistidos (id) ON DELETE CASCADE,
    oficina_id TEXT NOT NULL,
    PRIMARY KEY (assistido_id, oficina_id)
);
