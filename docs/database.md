# Banco de Dados — Vinyl Catalog

Guia para conectar e consultar o PostgreSQL que sustenta o Vinyl Catalog.

---

## Conexão

### Via Docker (sem instalar nada no host)

```bash
docker exec -it vinyl-catalog-db-1 psql -U postgres -d vinyl_catalog
```

### Via psql no host

Certifique-se de que o PostgreSQL está rodando e a porta 5432 está exposta:

```bash
psql -h localhost -p 5432 -U postgres -d vinyl_catalog
```

Senha padrão (conforme `.env`): `postgres`

### String de conexão DSN

```
postgresql://postgres:postgres@localhost:5432/vinyl_catalog?sslmode=disable
```

### Ferramentas GUI

| Ferramenta | Configuração |
|---|---|
| DBeaver / DataGrip | Host: `localhost`, Port: `5432`, DB: `vinyl_catalog`, User: `postgres`, Password: `postgres` |
| TablePlus | URL: `postgresql://postgres:postgres@localhost:5432/vinyl_catalog` |
| pgAdmin | Server: `localhost:5432`, Username: `postgres`, Password: `postgres` |

---

## Esquema

### Tabela `vinyls`

Armazena os discos de vinil do catálogo.

```sql
CREATE TABLE vinyls (
    id          UUID PRIMARY KEY,
    title       TEXT NOT NULL,
    artist      TEXT NOT NULL,
    year        BIGINT NOT NULL,
    genre       TEXT,
    label       TEXT,
    description TEXT,
    cover_url   TEXT,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ
);
```

### Tabela `tracks`

Faixas de cada disco. Relacionada a `vinyls` via `vinyl_id`.

```sql
CREATE TABLE tracks (
    id         UUID PRIMARY KEY,
    vinyl_id   TEXT NOT NULL,           -- UUID do disco pai
    title      TEXT NOT NULL,
    position   BIGINT,                  -- ordem da faixa no disco
    lyrics     TEXT,                    -- letra completa (pode ser longa)
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);

-- Índice para buscas por disco
CREATE INDEX idx_tracks_vinyl_id ON tracks (vinyl_id);
```

### Tabela `profiles`

Perfil único do administrador. Sempre tem exatamente uma linha.

```sql
CREATE TABLE profiles (
    id                  UUID PRIMARY KEY,
    username            TEXT NOT NULL UNIQUE,
    display_name        TEXT,
    bio                 TEXT,
    photo_url           TEXT,
    links               TEXT,   -- JSON array, ex: ["https://github.com/..."]
    preferred_genres    TEXT,   -- JSON array, ex: ["Jazz","Soul"]
    favorite_vinyl_ids  TEXT    -- IDs separados por vírgula
);

CREATE UNIQUE INDEX idx_profiles_username ON profiles (username);
```

---

## Consultas Úteis

### Discos

```sql
-- Listar todos os discos ordenados por artista
SELECT id, title, artist, year, genre, label
FROM vinyls
ORDER BY artist, year;

-- Buscar disco por título (case-insensitive)
SELECT * FROM vinyls
WHERE lower(title) LIKE lower('%thriller%');

-- Discos de um gênero específico
SELECT id, title, artist, year
FROM vinyls
WHERE genre = 'Jazz'
ORDER BY year;

-- Discos por ano (intervalo)
SELECT id, title, artist, year, genre
FROM vinyls
WHERE year BETWEEN 1960 AND 1980
ORDER BY year;

-- Discos com capa enviada
SELECT id, title, cover_url
FROM vinyls
WHERE cover_url IS NOT NULL AND cover_url <> '';

-- Discos sem capa
SELECT id, title, artist
FROM vinyls
WHERE cover_url IS NULL OR cover_url = '';

-- Total de discos no catálogo
SELECT count(*) AS total FROM vinyls;

-- Discos agrupados por gênero
SELECT genre, count(*) AS total
FROM vinyls
GROUP BY genre
ORDER BY total DESC;

-- Discos agrupados por década
SELECT (year / 10 * 10) AS decada, count(*) AS total
FROM vinyls
GROUP BY decada
ORDER BY decada;
```

### Faixas

```sql
-- Faixas de um disco específico (por ID)
SELECT position, title, length(lyrics) AS lyrics_chars
FROM tracks
WHERE vinyl_id = '2f8ff72d-3c5f-4238-9fd2-d60b847c7b8f'
ORDER BY position;

-- Faixas de todos os discos com nome do disco
SELECT v.title AS disco, t.position, t.title AS faixa
FROM tracks t
JOIN vinyls v ON v.id::text = t.vinyl_id
ORDER BY v.title, t.position;

-- Total de faixas por disco
SELECT v.title, count(t.id) AS faixas
FROM vinyls v
LEFT JOIN tracks t ON v.id::text = t.vinyl_id
GROUP BY v.id, v.title
ORDER BY faixas DESC;

-- Faixas que têm letra cadastrada
SELECT v.title AS disco, t.title AS faixa, length(t.lyrics) AS chars
FROM tracks t
JOIN vinyls v ON v.id::text = t.vinyl_id
WHERE t.lyrics IS NOT NULL AND t.lyrics <> ''
ORDER BY chars DESC;

-- Faixas sem letra
SELECT v.title AS disco, t.title AS faixa
FROM tracks t
JOIN vinyls v ON v.id::text = t.vinyl_id
WHERE t.lyrics IS NULL OR t.lyrics = '';

-- Total de faixas no catálogo
SELECT count(*) AS total FROM tracks;
```

### Perfil e Favoritos

```sql
-- Ver perfil completo
SELECT * FROM profiles;

-- Ver apenas favoritos (IDs separados por vírgula)
SELECT username, favorite_vinyl_ids
FROM profiles;

-- Discos favoritos com detalhes (expandindo a lista de IDs)
SELECT v.id, v.title, v.artist, v.year, v.genre
FROM vinyls v
JOIN profiles p ON p.favorite_vinyl_ids LIKE '%' || v.id::text || '%'
WHERE p.username = 'admin'
ORDER BY v.artist;

-- Gêneros preferidos do perfil
SELECT username, preferred_genres FROM profiles;

-- Links do perfil (JSON armazenado como texto)
SELECT username, links FROM profiles;
```

### Consultas de Manutenção

```sql
-- Discos criados nas últimas 24h
SELECT id, title, artist, created_at
FROM vinyls
WHERE created_at > now() - interval '24 hours'
ORDER BY created_at DESC;

-- Discos atualizados recentemente
SELECT id, title, updated_at
FROM vinyls
ORDER BY updated_at DESC
LIMIT 10;

-- Verificar integridade: faixas sem disco correspondente
SELECT t.id, t.vinyl_id, t.title
FROM tracks t
LEFT JOIN vinyls v ON v.id::text = t.vinyl_id
WHERE v.id IS NULL;

-- Tamanho das tabelas
SELECT
  relname AS tabela,
  pg_size_pretty(pg_total_relation_size(relid)) AS tamanho_total
FROM pg_catalog.pg_statio_user_tables
ORDER BY pg_total_relation_size(relid) DESC;

-- Tamanho total do banco
SELECT pg_size_pretty(pg_database_size('vinyl_catalog')) AS tamanho;
```

---

## Comandos psql Úteis

```sql
-- Listar tabelas
\dt

-- Descrever estrutura de uma tabela
\d vinyls
\d tracks
\d profiles

-- Ver tamanho das tabelas
\l+

-- Sair do psql
\q

-- Ver histórico de comandos
\s

-- Formatar saída em linhas (útil para registros com muitas colunas)
\x on
SELECT * FROM profiles;
\x off

-- Exportar resultado para CSV
\copy (SELECT * FROM vinyls ORDER BY artist) TO '/tmp/vinyls.csv' CSV HEADER;
```

---

## Backup e Restore

### Dump completo do banco

```bash
docker exec vinyl-catalog-db-1 pg_dump -U postgres vinyl_catalog > backup.sql
```

### Dump apenas dos dados (sem estrutura)

```bash
docker exec vinyl-catalog-db-1 pg_dump -U postgres --data-only vinyl_catalog > data.sql
```

### Restore

```bash
docker exec -i vinyl-catalog-db-1 psql -U postgres vinyl_catalog < backup.sql
```

### Dump em formato binário (recomendado para backups grandes)

```bash
docker exec vinyl-catalog-db-1 pg_dump -U postgres -Fc vinyl_catalog > backup.dump
docker exec -i vinyl-catalog-db-1 pg_restore -U postgres -d vinyl_catalog < backup.dump
```

---

## Configuração de Conexão (variáveis de ambiente)

| Variável | Padrão | Descrição |
|---|---|---|
| `DB_HOST` | `localhost` | Host do PostgreSQL |
| `DB_PORT` | `5432` | Porta |
| `DB_USER` | `postgres` | Usuário |
| `DB_PASSWORD` | `postgres` | Senha |
| `DB_NAME` | `vinyl_catalog` | Nome do banco |
| `DB_SSLMODE` | `disable` | Modo SSL (`require` em produção) |

Defina essas variáveis no arquivo `.env` na raiz do projeto.
