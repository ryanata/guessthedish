# Guess the Dish Architecture

## Deployment

Guess the Dish runs alongside other projects on the existing Azure
`Standard_D2ps_v6` VM in East US:

- Ubuntu 24.04 ARM64
- 2 Azure Cobalt 100 physical cores
- 8 GiB RAM
- Static public IP
- systemd-managed services

The Go server uses the standard library where practical and keeps external
dependencies minimal. It builds for ARM64 and AMD64 so it can move to another
provider. Each project on the VM has its own Unix user, directory, environment
file, localhost port, systemd unit, and resource limits.

```text
Browser
  React UI and answer search
          |
       HTTPS/WSS
          |
       Azure NSG
      TCP 80 and 443
          |
         Caddy
    TLS and host routing
          |
   127.0.0.1:<app port>
          |
  Application server
    Go
    HTTP and static assets
    WebSocket gateway
    Matchmaking and rooms
    Match state machines
    Bot engine
          |
    In-memory game state
    Private puzzle bundle
          |
       /metrics
          |
      Prometheus
          |
        Grafana
```

The Azure NSG allows public inbound traffic only on TCP ports 80 and 443. Caddy
is the only public-facing process. SSH is accessible through Tailscale;
application ports and data services remain bound to localhost.

## Components

| Component | Purpose | Location |
| --- | --- | --- |
| React/Vite client | Game UI, timers, and autocomplete | Browser; static assets served by the application server through Caddy |
| Caddy | TLS certificates, HTTPS, WebSocket proxying, and routing between projects by hostname | Shared systemd service on the VM |
| Go application server | HTTP API, static assets, WebSockets, and authoritative game rules, implemented with the standard library and minimal external dependencies | Dedicated systemd service on the VM |
| Match state machines | Own puzzle selection, reveal deadlines, guesses, scores, lockouts, and forfeits | Application process memory |
| Matchmaking and rooms | Quick Play queue, bot fallback, invite rooms, and reconnection reservations | Application process memory |
| Bot engine | Makes guesses from revealed clues using configured skill profiles | Application process |
| Public answer catalog | Canonical dish IDs, names, and aliases used for instant autocomplete | Downloaded to the browser |
| Private puzzle bundle | Maps puzzles to answers and contains ordered clues and content metadata | Server artifact only |
| Prometheus | Scrapes and retains application and VM metrics | Shared systemd service on the VM, bound to localhost |
| Node exporter | Reports VM CPU, memory, disk, and network metrics | Shared systemd service on the VM, bound to localhost |
| Grafana | Dashboards for game health and VM capacity | Shared systemd service on the VM, accessible through Tailscale only |
| CI/CD | Tests and builds ARM64 and AMD64 release artifacts, then deploys them | GitHub Actions and the VM |

## Real-Time Model

Each browser uses one WebSocket connection. The server controls all game state
and determines the first correct guess by server arrival order. Clients receive
absolute server deadlines. Clients periodically exchange timestamps with the
server to estimate clock offset and network delay, then animate countdowns
against server time. This keeps both clients on the same authoritative timeline
despite different network latency, without changing server arrival ordering for
guesses.

On reconnect, a client presents its opaque session token and receives a full
state snapshot. Its seat remains reserved for the ten seconds specified in the
game rules.

The answer catalog is safe to expose because it contains the universe of valid
dishes, not the answer to the current puzzle. Server-side autocomplete would be
equally enumerable while adding latency. Puzzle-to-answer mappings, future
clues, and bot decisions never enter browser assets. The server validates every
submitted dish ID and enforces the 300 ms wrong-answer lock and rate limits.

## State

Active sessions, queues, rooms, matches, and timers live only in the Go server's
memory. Restarting the service or VM clears this state, closes WebSocket
connections, and ends active matches. Puzzle content and application releases
remain durable outside the process.

## Observability

The application exposes a Prometheus `/metrics` endpoint on its localhost
listener. Prometheus scrapes it and node exporter, then Grafana reads
Prometheus to display dashboards. None of these endpoints require a public NSG
rule; Grafana is reached through Tailscale.

Metrics cover active connections, rooms and matches, queue depth, bot
fill rate, reconnects, command errors, event-loop or scheduler delay, process
memory, CPU, disk, and service restarts. Labels stay bounded: player, room,
match, dish, token, and IP values are never metric labels.

Prometheus uses a 15-second scrape interval, 14-day retention, and a disk-size
limit. Prometheus and Grafana receive systemd memory limits so monitoring cannot
starve live games. Grafana stores its small configuration database locally;
dashboards and alert rules are also kept in source control for recovery.

## Source And Content

The project uses exactly two repositories. The public repository contains the
application, protocol schemas, puzzle schema, validators, tests, and sample
puzzles. The private repository contains production recipes, aliases, ordered
clues, and content metadata.

CI checks out both repositories and combines the public application with the
private content into a server release. The browser receives only the public
answer catalog. Builds happen off the VM so one project's compilation cannot
interfere with live matches.

## VM Layout

```text
/srv/apps/guessthedish/releases/<release-id>/
/srv/apps/guessthedish/current -> releases/<release-id>
/etc/apps/guessthedish.env
/var/lib/guessthedish/
```

The service runs under a dedicated `guessthedish` user and listens only on its
assigned loopback port. systemd applies restart behavior, filesystem isolation,
and CPU/memory limits. Logs go to the systemd journal with bounded retention.

Deployments upload an immutable release, stop new matchmaking, allow active
matches a bounded drain period, switch the `current` symlink, restart only this
service, and roll back if readiness fails.
