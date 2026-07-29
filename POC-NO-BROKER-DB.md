# HyperFleet No-Broker DB-as-Queue PoC

## Repositories

| Component | Branch |
|---|---|
| [hyperfleet-api](https://github.com/rh-amarin/hyperfleet-api/tree/poc-no-broker-db) | `poc-no-broker-db` |
| [hyperfleet-adapter](https://github.com/rh-amarin/hyperfleet-adapter/tree/poc-no-broker-db) | `poc-no-broker-db` |
| [hyperfleet-infra](https://github.com/rh-amarin/hyperfleet-infra/tree/poc-no-broker-db) | `poc-no-broker-db` |

---

## Goals

This PoC evolves the `poc-no-broker` approach by replacing the direct HTTP push model with a **DB-as-queue** pattern. The API writes reconciliation messages to a PostgreSQL `messages` table, and adapters consume messages directly from that table. This retains all the benefits of removing the broker and sentinel while addressing the tight coupling and durability limitations of the HTTP push model.

Specific goals:

1. **Message durability** — if an adapter is down, messages accumulate in the database and are processed when the adapter restarts. No messages are lost.
2. **Decouple API from adapter availability** — the API no longer needs to know adapter URLs or wait for HTTP responses. It writes to the database and moves on.
3. **Concurrent processing** — adapters use a worker pool with configurable concurrency, replacing the unbounded goroutine-per-request model of the HTTP approach.
4. **No additional infrastructure** — messages live in the same PostgreSQL instance the API already uses. No broker, no sentinel, no new databases.
5. **Low-latency delivery** — PostgreSQL `LISTEN/NOTIFY` wakes adapters within milliseconds of a new message, with polling as a fallback.

---

## Architecture: Before and After

### Before — No-Broker / Direct HTTP Model (`poc-no-broker`)

```
  ┌──────────────────────────────────────────────────────────┐
  │                     HyperFleet API                        │
  │                                                           │
  │  Reconciler ──► HTTP POST /reconcile ──► Adapter :8082   │
  └──────────────────────────────────────────────────────────┘
```

The API reconciler directly HTTP-POSTs to each adapter's `/reconcile` endpoint. Adapters must be reachable at deploy time, the API blocks on HTTP responses, and messages are lost if an adapter is unavailable.

### After — No-Broker / DB-as-Queue Model (`poc-no-broker-db`)

```
  ┌──────────────────────────────────────────────────────────────────────┐
  │                         HyperFleet API                               │
  │                                                                      │
  │  Reconciler ──► INSERT INTO messages ──► NOTIFY 'messages'          │
  └──────────────────────────┬───────────────────────────────────────────┘
                             │
                    ┌────────▼────────┐
                    │   PostgreSQL    │
                    │  messages table │
                    └────────┬────────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
          ▼                  ▼                  ▼
    ┌───────────┐     ┌───────────┐      ┌───────────┐
    │cl-namespace│     │  cl-job   │      │cl-deploy. │   (+ more)
    │  adapter   │     │  adapter  │      │  adapter  │
    │            │     │           │      │           │
    │ LISTEN     │     │ LISTEN    │      │ LISTEN    │
    │ + poll 5s  │     │ + poll 5s │      │ + poll 5s │
    │ workers: 5 │     │ workers: 5│      │ workers: 5│
    └─────┬──────┘     └─────┬─────┘      └─────┬─────┘
          │                  │                   │
          └──────────────────┴───────────────────┘
                             │ POST /clusters/{id}/statuses
                             ▼
                     HyperFleet API (status write)
```

**Data flow for a CREATE:**

1. Client POSTs to `/clusters` → API writes resource to DB
2. Reconciler goroutine detects unreconciled resource on next poll cycle
3. Reconciler INSERTs one message per required adapter into the `messages` table
4. Reconciler issues `SELECT pg_notify('messages', '')` to wake listening adapters
5. Each adapter's consumer claims its messages via `FOR UPDATE SKIP LOCKED`
6. Worker pool processes messages concurrently, calling the executor
7. Completed messages are deleted from the table; failed messages are kept with error details
8. Each adapter POSTs status back to the API
9. API aggregates adapter statuses, sets `Reconciled=True` when all required adapters report success

---

## Messages Table Schema

```sql
CREATE TABLE messages (
    id            VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    adapter_name  VARCHAR(255) NOT NULL,
    kind          VARCHAR(100) NOT NULL,
    resource_id   VARCHAR(255) NOT NULL,
    payload       JSONB NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    claimed_at    TIMESTAMPTZ NULL,
    completed_at  TIMESTAMPTZ NULL,
    error_message TEXT NULL
);

-- Fast lookup for pending messages by adapter
CREATE INDEX idx_messages_adapter_status
    ON messages (adapter_name, status) WHERE status = 'pending';

-- Prevent duplicate pending/claimed messages for same resource+adapter
CREATE UNIQUE INDEX idx_messages_dedup
    ON messages (adapter_name, resource_id) WHERE status IN ('pending', 'claimed');
```

### Message Lifecycle

```
pending ──► claimed ──► deleted (on success)
                   └──► failed  (on error, kept for inspection)
```

### Deduplication

The unique partial index on `(adapter_name, resource_id) WHERE status IN ('pending', 'claimed')` ensures that only one in-flight message exists per adapter+resource pair. The reconciler uses `INSERT ... ON CONFLICT DO NOTHING` so repeated reconciler cycles are idempotent.

---

## Key Mechanisms

### API Reconciler (message producer)

The reconciler runs as a background goroutine inside the API process. On each tick:

1. Queries for resources whose last successful reconciliation is older than `stale_threshold`
2. For each stale resource, INSERTs a message per required adapter with `ON CONFLICT DO NOTHING`
3. After inserting, issues `SELECT pg_notify('messages', '')` to wake adapters

Uses `database/sql` directly (not GORM) for raw SQL inserts and NOTIFY.

### Adapter Consumer (message consumer)

Each adapter runs a `messagequeue.Consumer` that:

1. Opens a PostgreSQL `LISTEN` on the `messages` channel for instant wakeup
2. Polls every 5s as a fallback (configurable via `poll_interval`)
3. Claims messages in batches using:
   ```sql
   UPDATE messages SET status = 'claimed', claimed_at = NOW()
   WHERE id IN (
       SELECT id FROM messages
       WHERE adapter_name = $1 AND status = 'pending'
       ORDER BY created_at
       LIMIT $2
       FOR UPDATE SKIP LOCKED
   ) RETURNING *
   ```
4. Dispatches claimed messages to a bounded worker pool (default 5 workers)
5. On success: `DELETE FROM messages WHERE id = $1`
6. On failure: `UPDATE messages SET status = 'failed', error_message = $2`

`FOR UPDATE SKIP LOCKED` ensures multiple adapter replicas can claim concurrently without contention.

---

## Configuration

### API — Reconciler Settings

| Setting | Default | Env Override | Description |
|---|---|---|---|
| `reconciler.poll_interval` | `30s` | `RECONCILER_POLL_INTERVAL` | How often the reconciler scans for stale resources |
| `reconciler.stale_threshold` | `30m` | `RECONCILER_STALE_THRESHOLD` | Age at which a resource is considered stale |
| `reconciler.enabled` | `true` | — | Enable/disable the reconciler |

### API — Required Adapters

Adapter URLs are no longer needed. `required_adapters` is a `map[string]string` where only the key (adapter name) matters:

```yaml
required_adapters:
  cl-namespace: ""
  cl-job: ""
  cl-deployment: ""
  cl-maestro: ""
```

### Adapter — Message Queue Settings

| Setting | Default | Env Override | Description |
|---|---|---|---|
| `clients.message_queue.poll_interval` | `5s` | `HYPERFLEET_MESSAGE_QUEUE_POLL_INTERVAL` | Polling fallback interval |
| `clients.message_queue.workers` | `5` | `HYPERFLEET_MESSAGE_QUEUE_WORKERS` | Number of concurrent worker goroutines |
| `clients.message_queue.batch_size` | `10` | `HYPERFLEET_MESSAGE_QUEUE_BATCH_SIZE` | Max messages claimed per poll cycle |

### Adapter — Database Connection

Adapters connect to the same PostgreSQL instance as the API. Credentials are injected via the same Kubernetes secret (`hyperfleet-api-db-secrets`):

| Setting | Env Override |
|---|---|
| `clients.database.host` | `HYPERFLEET_DATABASE_HOST` |
| `clients.database.port` | `HYPERFLEET_DATABASE_PORT` |
| `clients.database.name` | `HYPERFLEET_DATABASE_NAME` |
| `clients.database.username` | `HYPERFLEET_DATABASE_USERNAME` |
| `clients.database.password` | `HYPERFLEET_DATABASE_PASSWORD` |
| `clients.database.ssl_mode` | `HYPERFLEET_DATABASE_SSL_MODE` |

---

## How to Deploy

### Prerequisites

| Tool | Version |
|---|---|
| [kubectl](https://kubernetes.io/docs/tasks/tools/) | >= 1.29 |
| [helm](https://helm.sh/) | >= 3.14 |
| [helmfile](https://helmfile.readthedocs.io/) | >= 0.162 |
| [podman](https://podman.io/) or docker | any recent |
| [go](https://go.dev/) | >= 1.25 |

### 1. Clone repos on the `poc-no-broker-db` branch

```bash
BRANCH=poc-no-broker-db

git clone -b $BRANCH https://github.com/rh-amarin/hyperfleet-api
git clone -b $BRANCH https://github.com/rh-amarin/hyperfleet-adapter
git clone -b $BRANCH https://github.com/rh-amarin/hyperfleet-infra
```

### 2. Build and push images

```bash
# Build for AMD64 (GKE) or ARM64 (kind on Apple Silicon)
cd hyperfleet-api
podman build --platform linux/amd64 -t quay.io/<you>/hyperfleet-api:poc-no-broker-db .
podman push quay.io/<you>/hyperfleet-api:poc-no-broker-db

cd ../hyperfleet-adapter
podman build --platform linux/amd64 -t quay.io/<you>/hyperfleet-adapter:poc-no-broker-db .
podman push quay.io/<you>/hyperfleet-adapter:poc-no-broker-db
```

### 3. Deploy with helmfile

```bash
cd ../hyperfleet-infra/helmfile

# GKE
NAMESPACE=hyperfleet-e2e-db \
  REGISTRY=quay.io \
  API_REPOSITORY=<you>/hyperfleet-api \
  API_IMAGE_TAG=poc-no-broker-db \
  ADAPTER_REPOSITORY=<you>/hyperfleet-adapter \
  ADAPTER_IMAGE_TAG=poc-no-broker-db \
  IMAGE_PULL_POLICY=Always \
  RUN_ID=manual-test \
  helmfile --environment e2e-gcp --kube-context <your-gke-context> apply

# kind
NAMESPACE=hyperfleet-db-queue \
  REGISTRY=quay.io \
  API_REPOSITORY=<you>/hyperfleet-api \
  API_IMAGE_TAG=poc-no-broker-db \
  ADAPTER_REPOSITORY=<you>/hyperfleet-adapter \
  ADAPTER_IMAGE_TAG=poc-no-broker-db \
  IMAGE_PULL_POLICY=Always \
  RUN_ID=manual-test \
  helmfile --environment e2e-kind --kube-context <your-kind-context> apply
```

### 4. Verify

```bash
# Check pods
kubectl -n <namespace> get pods

# Check adapter debug logs (polling + NOTIFY)
kubectl -n <namespace> logs -l app.kubernetes.io/name=hyperfleet-adapter -f | grep -E 'polling|claimed|NOTIFY'

# Check messages table
kubectl -n <namespace> exec <postgresql-pod> -- psql -U hyperfleet -d hyperfleet \
  -c "SELECT adapter_name, status, count(*) FROM messages GROUP BY adapter_name, status"

# Create a test cluster
kubectl -n <namespace> port-forward svc/hyperfleet-api 8000:8000 &
curl -s http://localhost:8000/api/hyperfleet/v1/clusters \
  -H 'Content-Type: application/json' \
  -d '{"kind":"Cluster","name":"test-cluster","spec":{"platform":{"type":"AWS"},"region":"us-east-1","version":"4.17.0"}}'

# Watch messages appear and get processed
watch -n1 'kubectl -n <namespace> exec <postgresql-pod> -- psql -U hyperfleet -d hyperfleet \
  -c "SELECT adapter_name, status, created_at, claimed_at FROM messages ORDER BY created_at DESC LIMIT 10"'
```

---

## Component Changes (from `poc-no-broker`)

### hyperfleet-api

| File | Change |
|---|---|
| `pkg/db/migrations/202607280001_add_messages.go` | **New.** Migration creating the `messages` table with dedup and status indexes |
| `pkg/db/migrations/migration_structs.go` | Added `addMessages()` to MigrationList |
| `pkg/reconciler/reconciler.go` | Replaced HTTP POST with `INSERT INTO messages ... ON CONFLICT DO NOTHING` + `pg_notify`. Removed `httpClient`, added `directDB` via `sessionFactory.DirectDB()` |
| `pkg/config/config.go` | Removed `HTTPTimeout` from `ReconcilerConfig` |
| `charts/templates/configmap.yaml` | Removed `http_timeout` from reconciler section |
| `charts/values.yaml` | Removed `http_timeout` from reconciler defaults |

### hyperfleet-adapter

| File | Change |
|---|---|
| `internal/messagequeue/consumer.go` | **New.** Worker pool consumer with LISTEN/NOTIFY + polling, `FOR UPDATE SKIP LOCKED` claiming, bounded concurrency |
| `internal/configloader/types.go` | Added `DatabaseConfig`, `MessageQueueConfig` structs and fields on `ClientsConfig` |
| `internal/configloader/viper_loader.go` | Added env var bindings for database and message queue config |
| `cmd/adapter/main.go` | Replaced HTTP reconcile server with DB message queue consumer. Added `database/sql`, `lib/pq` imports |
| `charts/templates/deployment.yaml` | Added database env vars from `database.secretName` secret |
| `charts/values.yaml` | Added `database.secretName` section. Removed port 8082 (reconcile) from `containerPorts` |

### hyperfleet-infra

| File | Change |
|---|---|
| `helmfile/helmfile.yaml.gotmpl` | Changed chart references from git URLs to local paths (`../../hyperfleet-api/charts`, `../../hyperfleet-adapter/charts`). Set `repositories: []` |
| `helmfile/values/base-api.yaml.gotmpl` | Changed `required_adapters` values from URLs to empty strings. Removed `http_timeout` |
| `helmfile/values/base-adapter.yaml.gotmpl` | Added `database.secretName: "hyperfleet-api-db-secrets"` |

---

## Key Design Decisions

| Decision | Rationale |
|---|---|
| DB-as-queue instead of HTTP push | Messages survive adapter downtime; API doesn't block on adapter availability |
| Same PostgreSQL instance | No new infrastructure; adapters share the API's existing database |
| `LISTEN/NOTIFY` + polling fallback | Sub-10ms delivery latency via NOTIFY; polling at 5s ensures no messages are missed if LISTEN drops |
| `FOR UPDATE SKIP LOCKED` | Multiple adapter replicas can claim messages concurrently without contention or deadlocks |
| `ON CONFLICT DO NOTHING` with dedup index | Reconciler cycles are idempotent; duplicate messages for the same resource+adapter are impossible |
| Delete on completion, keep on failure | Completed messages don't accumulate; failed messages remain for debugging |
| Worker pool (default 5) | Bounded concurrency replaces the unbounded goroutine-per-request model of the HTTP approach |
| Adapter names only (no URLs) | `required_adapters` values are empty strings; the API only needs adapter names to address messages |

---

## Comparison: `poc-no-broker` vs `poc-no-broker-db`

| Aspect | `poc-no-broker` (HTTP) | `poc-no-broker-db` (DB queue) |
|---|---|---|
| Transport | HTTP POST to adapter `/reconcile` | PostgreSQL `messages` table |
| Message durability | None (lost if adapter is down) | Full (persisted in DB) |
| API coupling | Must know adapter URLs | Only knows adapter names |
| Adapter concurrency | Unbounded goroutines | Bounded worker pool |
| Notification latency | Immediate (synchronous HTTP) | ~10ms (LISTEN/NOTIFY) |
| Infrastructure | None beyond API | None (same PostgreSQL) |
| Adapter availability | Required at reconcile time | Not required (messages queue up) |
| Deduplication | None (re-sends on every cycle) | Automatic via unique index |
