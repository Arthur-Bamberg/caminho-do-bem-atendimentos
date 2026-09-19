ALTER TABLE unidades ADD COLUMN IF NOT EXISTS ativo BOOLEAN NOT NULL DEFAULT true;

CREATE TABLE IF NOT EXISTS oficinas (
    id TEXT PRIMARY KEY,
    nome TEXT NOT NULL,
    ativo BOOLEAN NOT NULL DEFAULT true
);

INSERT INTO oficinas (id, nome, ativo) VALUES
    ('jiu-jitsu', 'Jiu-jitsu', true),
    ('npa-7-12', 'NPA (7 a 12 anos)', true),
    ('nacao-esporte', 'Nação Esporte', true),
    ('nacao-cultura', 'Nação Cultura', true),
    ('acessuas-trabalho', 'Acessuas Trabalho', true),
    ('juventude-na-mesa', 'Juventude na Mesa', true)
ON CONFLICT (id) DO NOTHING;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_oficina'
  ) THEN
    ALTER TABLE assistido_oficinas ADD CONSTRAINT fk_oficina
    FOREIGN KEY (oficina_id) REFERENCES oficinas(id);
  END IF;
END $$;

ALTER TABLE assistidos ADD COLUMN IF NOT EXISTS estrangeiro BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE assistidos ADD COLUMN IF NOT EXISTS pais_origem TEXT NOT NULL DEFAULT '';
