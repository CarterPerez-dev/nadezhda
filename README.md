<!-- ©AngelaMos | 2026 -->
<!-- README.md -->

```json
███╗   ██╗ █████╗ ██████╗ ███████╗███████╗██╗  ██╗██████╗  █████╗
████╗  ██║██╔══██╗██╔══██╗██╔════╝╚══███╔╝██║  ██║██╔══██╗██╔══██╗
██╔██╗ ██║███████║██║  ██║█████╗    ███╔╝ ███████║██║  ██║███████║
██║╚██╗██║██╔══██║██║  ██║██╔══╝   ███╔╝  ██╔══██║██║  ██║██╔══██║
██║ ╚████║██║  ██║██████╔╝███████╗███████╗██║  ██║██████╔╝██║  ██║
╚═╝  ╚═══╝╚═╝  ╚═╝╚═════╝ ╚══════╝╚══════╝╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝
```

[![Cybersecurity Projects](https://img.shields.io/badge/Cybersecurity--Projects-Project%20%2339-red?style=flat&logo=github)](https://github.com/CarterPerez-dev/Cybersecurity-Projects/tree/main/PROJECTS/intermediate/security-news-scraper)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev)
[![Keyless](https://img.shields.io/badge/enrichment-keyless-4B7BEC?style=flat)](https://www.cisa.gov/known-exploited-vulnerabilities-catalog)
[![Single static binary](https://img.shields.io/badge/binary-single%20static-6d4aff?style=flat)](https://pkg.go.dev/modernc.org/sqlite)
[![License: AGPLv3](https://img.shields.io/badge/License-AGPL_v3-purple.svg)](https://www.gnu.org/licenses/agpl-3.0)

> A keyless security-news and CVE intelligence engine in Go. Point it at a set of RSS and Atom feeds and it fetches them politely, clusters the same story as it breaks across outlets, extracts every CVE it mentions, enriches each one with authoritative exploit intelligence, ranks the whole set by what actually matters, and hands you a browsable dossier in your terminal. It ships as a single static binary over a local SQLite store, needs no API key to do any of it, and treats the news as the product with the CVE data as the intelligence that sharpens it.

<p align="center">
    <img src="assets/tui.png" alt="The nadezhda terminal UI: a ranked, color-coded dossier of security news, with per-story scores, source dots, and CVE severity" width="100%">
</p>

## Why aggregate security news

There is more security news published every day than any person can read, and almost all of it is noise relative to the one story that matters to you right now. The hard problem is not gathering headlines. It is deciding which three of the four hundred you should act on before lunch.

The signal is usually some combination of the same few factors. A vulnerability that many outlets picked up in the same few hours is trending for a reason. A CVE that CISA just added to its Known Exploited Vulnerabilities catalog is being used against real targets today. A flaw with a high EPSS score is one the rest of the world is about to start exploiting. Nadezhda computes that combination and sorts by it, so a CVSS 6.5 that is on the KEV catalog can outrank a CVSS 9.8 that nobody has touched.

The cost of missing that signal is not hypothetical. Log4Shell went from a GitHub issue to every security outlet on earth within hours and landed on the KEV catalog almost immediately; a tool that ranks by cross-outlet velocity plus KEV plus EPSS floats it to the top of the first scrape. Equifax fell to a known Apache Struts flaw that had a patch available for months, and the intelligence to prioritize it sat in public feeds the whole time, so the failure was one of triage, not of information. MOVEit was mass-exploited in a tight window where the advisories, the KEV listing, and the coverage all arrived together. Velocity was the tell.

## What it is

This is not a stub. It ingests real feeds, enriches real CVEs from keyless authoritative sources, clusters and ranks real coverage, and surfaces it four ways. Every capability below is exercised by unit tests, offline fixture-driven tests, and a live run against the real feeds.

**Ingestion**
- Seven seeded RSS/Atom feeds (Krebs on Security, The Hacker News, BleepingComputer, SecurityWeek, Dark Reading, The Register, CISA), fetched concurrently through a per-host rate limiter with an honest User-Agent
- Conditional GET: an unchanged feed answers `304 Not Modified` and costs almost nothing, and a retry honors `Retry-After` and never retries a timeout or a cancellation
- Fail-soft by construction: one broken feed is reported at the end and never aborts the run, and every article deduplicates at the database on both its canonical URL and a content hash

**Enrichment, keyless by default**
- CVE core from the CVE Program's cvelistV5 records: a CVSS score resolved by a fixed version precedence (v4.0, then v3.1, then v3.0, then v2.0) read across both the vendor and the CISA-ADP containers, plus CWE and description
- CISA KEV membership, with the `knownRansomwareCampaignUse` string mapped explicitly to a real ransomware signal rather than treated as a boolean
- FIRST EPSS probability and percentile, parsed from the quoted strings the API actually returns so the score never silently reads as zero
- A TTL cache with a separate, shorter negative TTL, so a not-yet-published CVE is not re-fetched every run; NVD 2.0 is supported as an optional key-gated booster, never a requirement

**Clustering and ranking**
- Connected-components (union-find) clustering: two items join when they share a CVE, or when they come from different outlets and their title token sets clear a Jaccard threshold inside a time window
- A pure, deterministic, news-first weighted score: recency, cross-outlet velocity, source trust, and a keyword watchlist carry 70 percent of the weight, with KEV, CVSS, and EPSS as the supporting 30, so the same corpus always sorts the same way and is pinned by golden-order tests

**Surfaces**
- A colorful bubbletea terminal UI: a ranked list of dossiers with severity-reactive rows and KEV chips, a scrollable detail view with every outlet link and the full CVE card, and `o` to open a story in your browser
- A Markdown or JSON digest of the top-ranked clusters, for feeding a newsletter, a report, or another tool
- A single-CVE lookup that shows the enriched record and every stored story that mentions it

**Content ideation, optional AI, off by default**
- Turn ranked clusters into summaries and content angles through one `Provider` interface: a local Qwen model via Ollama (keyless default), or Claude, OpenAI, or Gemini behind a single pasted key
- A re-runnable in-binary setup wizard writes keys to a `0600` credentials file behind a name allowlist, read from the environment and never logged

**Watch daemon, optional**
- A long-lived scheduler that re-ingests on a timer and posts genuinely new, high-signal stories to Slack, Discord, or any JSON webhook
- New is measured on fetch time, not publish time, so a backfilled advisory fetched today still alerts exactly once; the daemon shuts down cleanly on Ctrl-C or SIGTERM and never crashes on a transient feed or network error

## Quick Start

```bash
curl -fsSL https://angelamos.com/nadezhda/install.sh | bash
```

One command, zero further steps: it grabs a prebuilt binary for your platform (no Go toolchain needed), drops it on your `PATH`, and leaves `nadezhda` runnable by name. Then pull the news and browse it:

```bash
nadezhda scrape              # pull every enabled feed, cluster, and enrich CVEs
nadezhda tui                 # browse the ranked dossier (? for keys, o opens a story)
nadezhda digest --top 20     # render a ranked Markdown digest (or --format json)
nadezhda cve CVE-2021-44228  # one enriched CVE and the stories that mention it
```

The default flow is `scrape` then `tui`; scrape already runs enrichment best-effort, so there is nothing else to wire up. A scrape prints a per-source table and a one-line summary:

```
SOURCE             STATUS   PARSED   NEW   DUP   CVE   ERR
krebs              ok       10       3     7     1     0
bleepingcomputer   ok       18       6     12    4     0
cisa               304      -        -     -     -     -
...
12 new, 45 duplicate, 8 CVE refs across 7 sources (0 failed)
132 clusters (14 multi-source, largest 5)
enriched 80/80 CVEs (14 KEV, 0 not found)
```

The optional layers are one command each:

```bash
nadezhda ai                  # set up AI ideation: paste one key, or point at a local Ollama
nadezhda watch --interval 1h # run as a daemon, alerting new high-signal stories to a webhook
```

Prefer the Go toolchain? A real `go install github.com/CarterPerez-dev/nadezhda/cmd/nadezhda@latest` works too, and `just build` builds from a checkout. Building from source needs Go 1.25+, fetched automatically if you are on an older Go.

> [!TIP]
> This project uses [`just`](https://github.com/casey/just) as a command runner. Type `just` to see every recipe.
>
> Install: `curl -sSf https://just.systems/install.sh | bash -s -- --to ~/.local/bin`

## Learn

This project ships a full teaching track. Read it in order, or jump to what you need.

| Doc | What it covers |
|-----|----------------|
| [`learn/00-OVERVIEW.md`](learn/00-OVERVIEW.md) | What the tool is, prerequisites, the project layout, and a quick tour |
| [`learn/01-CONCEPTS.md`](learn/01-CONCEPTS.md) | The CVE intelligence stack, clustering as a velocity signal, and the ranking model, grounded in real incidents |
| [`learn/02-ARCHITECTURE.md`](learn/02-ARCHITECTURE.md) | The pipeline, the package map, the data model, and the design decisions |
| [`learn/03-IMPLEMENTATION.md`](learn/03-IMPLEMENTATION.md) | A code walkthrough from a feed item to a ranked, enriched cluster |
| [`learn/04-CHALLENGES.md`](learn/04-CHALLENGES.md) | Extension ideas, from adding a source to replacing the clustering with LSH |

## Architecture

One Go module, no service boundary, no message queue, and no external database. The `scrape` command runs the whole left-to-right flow once; the `watch` daemon runs it on a timer. Everything in `internal` exists to feed one function, `ingestAndCluster`, and everything downstream exists to rank and present what it produced.

```
        sources.yaml (embedded default, or a user --config file)
                        │
                        ▼
   ┌──────────────────────────────────────────────┐
   │  fetch     N workers, per-host rate limit,    │  internal/fetch
   │            conditional GET (ETag / 304)       │
   └───────────────────────┬──────────────────────┘
                           │  raw feed bytes
                           ▼
   ┌──────────────────────────────────────────────┐
   │  parse → normalize → ingest                   │  parse / normalize / ingest
   │  RSS/Atom, canonical URL, hash, fail-soft     │
   └───────────────────────┬──────────────────────┘
             CVE regex      │      union-find
                 ┌──────────┴──────────┐
                 ▼                     ▼
   ┌──────────────────────────────────────────────┐
   │  enrich    CVE list / KEV / EPSS, cached,     │  enrich / cve
   │            keyless                            │
   └───────────────────────┬──────────────────────┘
                           ▼
   ┌──────────────────────────────────────────────┐
   │  rank      deterministic, news-first score    │  internal/rank
   └───────────────────────┬──────────────────────┘
              ┌────────────┼────────────┬───────────┐
              ▼            ▼            ▼           ▼
           digest         tui         ideate      watch
          (export)     (browse)       (AI)       (daemon)
```

**Design decisions:** the store is `modernc.org/sqlite`, a pure-Go SQLite with no CGO, so the four-platform release binaries fall out of one build and the core needs no database server. Enrichment folds into `scrape` because the keyless CVE source is unthrottled, and it runs best-effort under a timeout so the news never blocks or fails on a slow CVE lookup. The `watch` daemon is a pure scheduler that imports nothing from the store or the pipeline; the concrete work is injected as a closure, so the loop, the graceful shutdown, and the fail-soft behavior are all unit-tested with a fake clock and no network. Alerts use a fetch-time watermark, not publish time, so an advisory published last week but fetched today is correctly new to you.

## Build and Test

```bash
just build       # -> ./nadezhda   (or: go build -o nadezhda ./cmd/nadezhda)
just test        # go test ./...   (18 packages)
just ollama-up   # stand up the local Qwen runtime in Docker (optional AI)
```

Ranking is a pure function pinned to golden-order tests: fixed inputs produce one exact ordering. The keyless CVE client is pinned to a known-answer fixture, Log4Shell (CVE-2021-44228): the CVSS 10.0 score is read out of the CISA-ADP container because the vendor container holds only a placeholder, the weakness resolves to CWE-502, it is KEV-listed, and its EPSS score sits near the top of the scale. The `knownRansomwareCampaignUse` string mapping and the quoted-string EPSS parsing each have a regression test, because both are the quiet kind of bug that decodes wrong with no error. The watch scheduler is tested with a fake ticker a test drives by hand, so its behavior is verified with no clock, no network, and no database.

## Project Structure

```
security-news-scraper/
├── cmd/nadezhda/          # the CLI: one file per command + the shared pipeline seam
│   ├── scrape.go          # the whole tool in miniature: runScrape
│   ├── pipeline.go        # ingestAndCluster, the seam shared by scrape and watch
│   ├── tui.go             # digest.go ideate.go watch.go cve.go list.go sources.go ai.go
│   └── main.go            # root.go, the cobra command tree
├── internal/
│   ├── fetch/             # concurrent, rate-limited, conditional HTTP (+ an unused robots gate)
│   ├── parse/             # RSS/Atom via gofeed, HTML fallback via goquery
│   ├── normalize/         # canonical URLs, content hashing, time parsing
│   ├── ingest/            # the fan-out orchestrator, fail-soft per source
│   ├── cluster/           # union-find clustering by title similarity and shared CVE
│   ├── cve/               # keyless clients: cvelistV5, CISA KEV, FIRST EPSS (+ optional NVD)
│   ├── enrich/            # enrichment orchestration and the TTL cache
│   ├── rank/              # the pure, deterministic weighted score
│   ├── store/             # SQLite: connection, migrations, and every typed query
│   ├── ai/                # the opt-in ideation layer, four providers behind one interface
│   ├── setup/             # the in-binary credential wizard
│   ├── watch/             # the daemon scheduler, decoupled from the work it runs
│   ├── tui/               # the bubbletea terminal browser
│   ├── export/            # the Markdown and JSON digest renderers
│   ├── source/            # the source registry and the embedded sources.yaml
│   ├── config/            # every tunable, with defaults and validation
│   └── version/           # the ldflags-injected version string
├── testdata/              # captured feed and API fixtures, driving the offline tests
├── install.sh             # the one-shot curl-able installer
├── learn/                 # the teaching track (public)
└── justfile               # every recipe
```

## License

[AGPL 3.0](LICENSE).
