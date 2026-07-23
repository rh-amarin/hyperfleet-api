# HyperFleet No-Broker PoC

## Repositories

| Component | Branch |
|---|---|
| [hyperfleet-api](https://github.com/rh-amarin/hyperfleet-api/tree/poc-no-broker) | `poc-no-broker` |
| [hyperfleet-adapter](https://github.com/rh-amarin/hyperfleet-adapter/tree/poc-no-broker) | `poc-no-broker` |
| [hyperfleet-infra](https://github.com/rh-amarin/hyperfleet-infra/tree/poc-no-broker) | `poc-no-broker` |
| [hyperfleet-e2e](https://github.com/rh-amarin/hyperfleet-e2e/tree/poc-no-broker) | `poc-no-broker` |

---

## Goals

The primary goal of this PoC is to eliminate the message broker (GCP Pub/Sub / RabbitMQ) and the sentinel component from the HyperFleet architecture, replacing the indirect pub/sub fan-out with a direct HTTP push model driven by a reconciler loop embedded in the API.

Specific goals:

1. **Remove operational complexity** — no broker infrastructure to provision, secure, monitor, or tune per environment.
2. **Remove the sentinel component** — a dedicated process whose only job was to bridge the API database to the broker; its responsibilities move into the API itself.
3. **Simplify adapter deployment** — adapters no longer need broker credentials, subscription IDs, or topic names configured at deploy time.
4. **Make adapter URLs first-class** — `required_adapters` changes from a list of names to a name→URL map, letting the API know exactly where to reach each adapter without any service-discovery indirection.
5. **Maintain the same E2E guarantees** — clusters still reach `Reconciled=True` only when all required adapters report success; adapter failures still surface as conditions; finalization still blocks hard-delete until every adapter finalizes.

---

## Architecture: Before and After

### Before — Broker / Sentinel Model

```
  ┌────────────────────────────────────────────────────────────────┐
  │                        HyperFleet API                          │
  │  POST /clusters  ──► DB write                                  │
  └───────────────────────────┬────────────────────────────────────┘
                              │ DB
                              │
  ┌───────────────────────────▼────────────────────────────────────┐
  │                    hyperfleet-sentinel                         │
  │  polls API for new/updated resources                           │
  │  publishes event to broker topic                               │
  └───────────────────────────┬────────────────────────────────────┘
                              │ publish
                              ▼
              ┌───────────────────────────────┐
              │         Message Broker        │
              │  GCP Pub/Sub  or  RabbitMQ    │
              │                               │
              │  topic: {ns}-clusters         │
              │  dlq:   {ns}-clusters-dlq     │
              └───────┬───────────────────────┘
                      │ subscribe (one queue per adapter)
          ┌───────────┼───────────┬───────────────┐
          │           │           │               │
          ▼           ▼           ▼               ▼
    ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
    │cl-namesp.│ │ cl-job   │ │cl-deploy.│ │cl-maestro│
    │ adapter  │ │ adapter  │ │ adapter  │ │ adapter  │
    └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘
         │            │            │             │
         └────────────┴────────────┴─────────────┘
                              │ POST /clusters/{id}/statuses
                              ▼
                      HyperFleet API (status write)
```

**Data flow for a CREATE:**

1. Client POSTs to `/clusters` → API writes to DB
2. Sentinel polls DB, detects new record, publishes event to broker topic
3. Each adapter consumes from its dedicated subscription queue
4. Each adapter processes the event and POSTs status back to the API
5. API aggregates adapter statuses, sets `Reconciled=True` when all required adapters report success

**Required adapters config (entity descriptor):**

```yaml
required_adapters:
  - cl-namespace
  - cl-job
  - cl-deployment
  - cl-maestro
```

---

### After — No-Broker / Direct HTTP Model

```
  ┌────────────────────────────────────────────────────────────────────────────────┐
  │                            HyperFleet API                                      │
  │                                                                                │
  │  POST /clusters  ──► DB write                                                  │
  │                                                                                │
  │  ┌──────────────────────────────────────────────────────────────────────────┐  │
  │  │  Background Reconciler Goroutine                                         │  │
  │  │  • polls DB every 30s (configurable)                                    │  │
  │  │  • finds stale/unreconciled resources                                   │  │
  │  │  • POSTs plain JSON to each adapter URL in parallel                     │  │
  │  └──────────────────────────────────────────────────────────────────────────┘  │
  └───────┬──────────────────────────────────────────────────────────────────┬─────┘
          │ POST /reconcile (plain JSON)                    POST /clusters/{id}/statuses
          │                                                                  ▲
          ├───────────────────────────┐                                      │
          │                           │                                      │
          ▼                           ▼                                      │
    ┌─────────────────────┐   ┌─────────────────────┐                       │
    │  cl-namespace       │   │  cl-job             │   (+ more adapters)   │
    │  adapter :8082      │   │  adapter :8082      │                       │
    └──────────┬──────────┘   └──────────┬──────────┘                       │
               └─────────────────────────┴───────────────────────────────────┘
```

**Data flow for a CREATE:**

1. Client POSTs to `/clusters` → API writes to DB
2. Reconciler goroutine (inside API) picks up unreconciled resource on next poll cycle
3. API POSTs reconcile payload directly to each adapter's `/reconcile` endpoint in parallel
4. Each adapter processes and POSTs status back to the API
5. API aggregates adapter statuses, sets `Reconciled=True` when all required adapters report success

**Required adapters config (entity descriptor):**

```yaml
required_adapters:
  cl-namespace:  "http://cl-namespace.hyperfleet.svc.cluster.local:8082"
  cl-job:        "http://cl-job.hyperfleet.svc.cluster.local:8082"
  cl-deployment: "http://cl-deployment.hyperfleet.svc.cluster.local:8082"
  cl-maestro:    "http://cl-maestro.hyperfleet.svc.cluster.local:8082"
```

---

## How to Reproduce (kind)

This section walks through standing up the full no-broker stack on a local kind cluster.

### Prerequisites

| Tool | Version |
|---|---|
| [kind](https://kind.sigs.k8s.io/) | ≥ 0.23 |
| [kubectl](https://kubernetes.io/docs/tasks/tools/) | ≥ 1.29 |
| [helm](https://helm.sh/) | ≥ 3.14 |
| [helmfile](https://helmfile.readthedocs.io/) | ≥ 0.162 |
| [podman](https://podman.io/) or docker | any recent |
| [go](https://go.dev/) | ≥ 1.22 (to build images locally) |

### 1. Clone all four repos on the `poc-no-broker` branch

```bash
BRANCH=poc-no-broker

git clone -b $BRANCH https://github.com/rh-amarin/hyperfleet-api
git clone -b $BRANCH https://github.com/rh-amarin/hyperfleet-adapter
git clone -b $BRANCH https://github.com/rh-amarin/hyperfleet-infra
git clone -b $BRANCH https://github.com/rh-amarin/hyperfleet-e2e
```

### 2. Place the repos so infra can find them

The infra `kind-build-images.sh` script builds images from local source. By default it looks for sibling directories:

```
workspaces/
  hyperfleet-api/
  hyperfleet-adapter/
  hyperfleet-infra/       ← run make commands from here
  hyperfleet-e2e/
```

### 3. Create the kind cluster

```bash
cd hyperfleet-infra
make create-kind-cluster
```

This creates a cluster named `kind-poc` (or `kind` if `KIND_CLUSTER_NAME` is unset) and exports its kubeconfig.

### 4. Build and load images

```bash
# From hyperfleet-infra
make kind-build-images
```

This builds `hyperfleet-api` and `hyperfleet-adapter` images locally and loads them into the kind cluster via `kind load image-archive`. The sentinel image is **not** built in this PoC.

> To skip rebuilding and use pre-existing local images, set `BUILD_IMAGES=false` in `env.kind`.

### 5. Install Maestro (required for the `cl-maestro` adapter)

```bash
make install-maestro-all
```

This installs the Maestro server and agent, then registers a consumer named `cluster1`.

### 6. Configure environment

Copy the kind env file and set required variables:

```bash
cp env.kind .env.local   # optional local override
```

Key variables (all have defaults in `env.kind`):

| Variable | Default | Description |
|---|---|---|
| `HELMFILE_ENV` | — | Set to `kind` or `e2e-kind` |
| `NAMESPACE` | `hyperfleet-local` | Kubernetes namespace |
| `API_IMAGE_TAG` | `local` | Tag for the locally built API image |
| `ADAPTER_IMAGE_TAG` | `local` | Tag for the locally built adapter image |
| `RECONCILER_POLL_INTERVAL` | `5s` | How often the reconciler polls the DB |
| `RECONCILER_STALE_THRESHOLD` | `30m` | Age at which a resource is considered stale |
| `RECONCILER_HTTP_TIMEOUT` | `5s` | Timeout for adapter HTTP calls |

### 7. Deploy HyperFleet

```bash
# Install helm repos, then deploy API + adapters in one shot
HELMFILE_ENV=kind make install-repos install-hyperfleet
```

Or step by step:

```bash
HELMFILE_ENV=kind make install-repos
HELMFILE_ENV=kind make install-api
HELMFILE_ENV=kind make install-adapters
```

### 8. Verify the stack

```bash
kubectl -n hyperfleet-local get pods
```

Expected pods (no sentinel, no broker):

```
hyperfleet-api-*          Running
cl-namespace-*            Running
cl-job-*                  Running
cl-deployment-*           Running
cl-maestro-*              Running
np-configmap-*            Running
```

Tail the API logs to watch the reconciler:

```bash
kubectl -n hyperfleet-local logs -l app.kubernetes.io/name=hyperfleet-api -f | grep reconciler
```

### 9. Run E2E tests (optional)

```bash
cd ../hyperfleet-e2e

# Point at the API — get the ClusterIP or port-forward first
kubectl -n hyperfleet-local port-forward svc/hyperfleet-api 8000:8000 &

HYPERFLEET_API_URL=http://localhost:8000 \
  make e2e-ci GINKGO_LABEL_FILTER=tier0
```

### 10. Tear down

```bash
cd ../hyperfleet-infra
HELMFILE_ENV=kind make local-down-kind
```

---

## Component Changes

### hyperfleet-api

**What changed:**

- Added a **background reconciler goroutine** that starts at API boot time and runs until shutdown.
- The reconciler polls the database on a configurable interval (`poll_interval`, default `30s`) for resources that are stale (last reconciliation older than `stale_threshold`, default `30m`) or have never been reconciled.
- For each stale resource, the reconciler POSTs a plain JSON reconcile payload to every adapter URL listed in the entity's `required_adapters` map, concurrently.
- Added an HTTP client with configurable timeout (`http_timeout`, default `5s`) for adapter calls.
- Fixed a `context.Background()` bug where the reconciler was not inheriting the server's shutdown context, causing goroutine leaks on graceful stop.
- `required_adapters` in entity descriptors changed from `[]string` → `map[string]string` (name → URL).

**New Helm values:**

```yaml
reconciler:
  enabled: true
  poll_interval: 30s
  stale_threshold: 30m
  http_timeout: 5s
```

**Removed:**

- All broker client configuration (`broker.type`, `broker.googlepubsub.*`, `broker.rabbitmq.*`)

---

### hyperfleet-adapter

**What changed:**

- Exposed a new **`POST /reconcile`** HTTP endpoint on port `8082`.
- The adapter now acts as an HTTP server waiting for push from the API, instead of a subscriber pulling from a broker queue.
- Removed all broker client code (Pub/Sub subscriber, RabbitMQ consumer, subscription lifecycle management).

**Removed:**

- All broker configuration (`broker.type`, `googlepubsub.*`, `rabbitmq.*`)
- Broker Helm values (`broker.type`, `broker.googlepubsub.*`, `broker.rabbitmq.*`)

---

### hyperfleet-sentinel

**Status: removed entirely.**

The sentinel's only job was to detect DB changes and publish events to the broker. That responsibility now lives in the API's reconciler goroutine. The sentinel repository and Helm chart remain on disk but are not deployed in the no-broker architecture.

---

### hyperfleet-infra

**What changed:**

- **`helmfile/helmfile.yaml.gotmpl`**:
  - Removed `hyperfleet-sentinel` Helm repository registration
  - Removed `sentinel.chartRef` from the `charts:` block
  - Removed `brokerType` from all 4 environment definitions
  - Removed all `sentinel-configs` includes
  - Removed the `{{ range .Values.sentinels }}` release block
  - Removed the RabbitMQ conditional release block

- **`helmfile/values/base-api.yaml.gotmpl`**:
  - Changed `required_adapters` rendering from a list to a map (name→URL):

    ```gotmpl
    required_adapters:
      {{ .name }}: "http://{{ .name }}.{{ $.Values.namespace }}.svc.cluster.local:8082"
    ```

  - Added reconciler config section with env var overrides

- **`helmfile/values/base-adapter.yaml.gotmpl`**: Removed entire `broker:` stanza

- **`helmfile/values/base-sentinel.yaml.gotmpl`**: Deleted

- **`helmfile/environments/*/sentinel-configs.yaml[.gotmpl]`**: Deleted (4 files)

- **`helmfile/environments/e2e-gcp/adapter-configs.yaml.gotmpl`** and **`e2e-kind`**:
  - Removed all `broker:` blocks from each adapter config

- **`helmfile/configs/*/adapters/*/adapter-config.yaml`** (8 files): Removed `clients.broker` section

- **`terraform/helm-values-files.tf`**: Removed `sentinel_values` local and `local_file.sentinel_values` resource

- **`terraform/outputs.tf`**: Removed sentinels block from `helm_values` output

- **`scripts/kind-build-images.sh`**: Removed sentinel; `COMPONENTS=(api sentinel adapter)` → `COMPONENTS=(api adapter)`

- **`scripts/generate-rabbitmq-values.sh`**: Deleted

- **`Makefile`**: Removed `generate-rabbitmq-values`, `install-sentinels`, `uninstall-sentinels` targets and RabbitMQ variables

- **`env.kind`** and **`env.gcp`**: Removed `RABBITMQ_URL`, `SENTINEL_*` variables

---

### hyperfleet-e2e

**What changed:**

- **`pkg/config/config.go`**: Removed `BrokerType` field and validation

- **`pkg/config/defaults.go`**: `DefaultClusterAdapters` and `DefaultNodePoolAdapters` changed from `[]string` → `map[string]string` (name→URL); removed `DefaultBrokerType`

- **`pkg/helper/api_config.go`**: `UpgradeAPIRequiredAdapters`, `RestoreAPIRequiredAdaptersWithRetry`, `GetAPIRequiredClusterAdapters`, `patchEntityRequiredAdapters` all updated to use `map[string]string`

- **`pkg/helper/matchers.go`**: `HaveAllAdaptersWithCondition` and `HaveAllAdaptersAtGeneration` updated to accept `map[string]string`

- **`pkg/helper/adapter.go`**: Removed `PurgeAdapterQueue`, `purgeRabbitMQQueue`, `DeletePubSubResourcesForAdapter`, `DeletePubSubSubscription`, `DeletePubSubTopic` and all GCP Pub/Sub imports

- **`pkg/helper/cleanup.go`**: Removed `SweepPubsubTestAdapterResources` and broker-type-conditional cleanup

- **`pkg/helper/constants.go`**: Removed `SentinelClustersRelease` and `SentinelNodePoolsRelease`

- **E2E test files** (`crash_recovery.go`, `stuck_deletion.go`, `adapter_failure.go`, `adapter_failover.go`, `adapter_with_maestro.go`, `delete_edge_cases.go`, `force_delete.go`):
  - Removed all `if h.Cfg.BrokerType == "googlepubsub" { DeletePubSubResourcesForAdapter(...) }` cleanup blocks
  - Removed all `PurgeAdapterQueue` calls
  - Fixed `updatedAdapters` construction from `[]string` append to `map[string]string` copy+add
  - Removed sentinel scale-down from `delete_edge_cases.go`
  - Fixed `Adapters.Cluster[0]` map index in `force_delete.go`

- **`env/env.ci`**: Removed `SENTINEL_IMAGE_REPO`, `SENTINEL_IMAGE_TAG`, `BROKER_TYPE`, `SENTINEL_BROKER_TYPE`, `SENTINEL_CHART_*`

- **`configs/config.yaml`**: `adapters.cluster` and `adapters.nodepool` changed from list to map format

- **`testdata/adapter-configs/*/values.yaml`** (8 files): Removed `broker:` Helm values block

- **`testdata/adapter-configs/*/adapter-config.yaml`** (8 files): Removed `clients.broker` section

---

## Key Design Decisions

| Decision | Rationale |
|---|---|
| Reconciler lives in the API, not a new sidecar | Keeps the deployment surface small; avoids a new component with its own lifecycle |
| Push (API→adapter) not pull (adapter→broker) | Eliminates broker infrastructure; adapter availability is observable immediately via HTTP response |
| `required_adapters` as map[name→URL] | URL is the authoritative address; name stays as the correlation key for status reporting |
| Poll-based reconciler with stale threshold | Handles restarts, missed events, and eventual consistency without exactly-once delivery guarantees |
| Adapter URL = `http://{name}.{namespace}.svc.cluster.local:8082` | Derived deterministically from adapter name and namespace; no service discovery layer needed |
