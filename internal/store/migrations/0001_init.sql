-- ©AngelaMos | 2026
-- 0001_init.sql

CREATE TABLE sources (
    id      INTEGER PRIMARY KEY,
    name    TEXT NOT NULL UNIQUE,
    title   TEXT NOT NULL DEFAULT '',
    url     TEXT NOT NULL,
    type    TEXT NOT NULL,
    weight  REAL NOT NULL DEFAULT 1.0,
    tags    TEXT NOT NULL DEFAULT '',
    enabled INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE fetch_state (
    source_id     INTEGER PRIMARY KEY REFERENCES sources(id) ON DELETE CASCADE,
    etag          TEXT NOT NULL DEFAULT '',
    last_modified TEXT NOT NULL DEFAULT '',
    last_fetched  INTEGER NOT NULL DEFAULT 0,
    last_status   INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE articles (
    id            INTEGER PRIMARY KEY,
    source_id     INTEGER NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    canonical_url TEXT NOT NULL UNIQUE,
    content_hash  TEXT NOT NULL UNIQUE,
    title         TEXT NOT NULL DEFAULT '',
    summary       TEXT NOT NULL DEFAULT '',
    body          TEXT NOT NULL DEFAULT '',
    author        TEXT NOT NULL DEFAULT '',
    published_at  INTEGER NOT NULL DEFAULT 0,
    fetched_at    INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_articles_published ON articles(published_at);
CREATE INDEX idx_articles_source ON articles(source_id);

CREATE TABLE cves (
    id              TEXT PRIMARY KEY,
    description     TEXT NOT NULL DEFAULT '',
    cvss_score      REAL,
    cvss_version    TEXT NOT NULL DEFAULT '',
    cvss_severity   TEXT NOT NULL DEFAULT '',
    cvss_vector     TEXT NOT NULL DEFAULT '',
    cwe             TEXT NOT NULL DEFAULT '',
    is_kev          INTEGER NOT NULL DEFAULT 0,
    kev_date_added  TEXT NOT NULL DEFAULT '',
    kev_ransomware  INTEGER NOT NULL DEFAULT 0,
    epss            REAL,
    epss_percentile REAL,
    nvd_published   TEXT NOT NULL DEFAULT '',
    nvd_modified    TEXT NOT NULL DEFAULT '',
    enriched_at     INTEGER NOT NULL DEFAULT 0,
    enrich_status   TEXT NOT NULL DEFAULT ''
);

CREATE TABLE article_cves (
    article_id INTEGER NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    cve_id     TEXT NOT NULL REFERENCES cves(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, cve_id)
);

CREATE TABLE clusters (
    id         INTEGER PRIMARY KEY,
    cluster_key TEXT NOT NULL UNIQUE,
    first_seen INTEGER NOT NULL DEFAULT 0,
    last_seen  INTEGER NOT NULL DEFAULT 0,
    size       INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE cluster_members (
    cluster_id INTEGER NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    article_id INTEGER NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    PRIMARY KEY (cluster_id, article_id)
);

CREATE TABLE ai_notes (
    id          INTEGER PRIMARY KEY,
    cluster_id  INTEGER NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    provider    TEXT NOT NULL DEFAULT '',
    summary     TEXT NOT NULL DEFAULT '',
    why         TEXT NOT NULL DEFAULT '',
    angles_json TEXT NOT NULL DEFAULT '',
    format      TEXT NOT NULL DEFAULT '',
    created_at  INTEGER NOT NULL DEFAULT 0
);
