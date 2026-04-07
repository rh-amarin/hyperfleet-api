# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0](https://github.com/openshift-hyperfleet/hyperfleet-api/compare/v0.1.1...v0.2.0) - 2026-04-03

### Added

- Aggregation logic for resource data **PR:** [#91](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/91) **Jira:** [HYPERFLEET-774](https://issues.redhat.com/browse/HYPERFLEET-774)
  - Keeps aggregated condition timestamps aligned with adapter generation and rejects invalid condition status values.
- Version subcommand to CLI **PR:** [#84](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/84) **Jira:** [HYPERFLEET-727](https://issues.redhat.com/browse/HYPERFLEET-727)
  - Exposes build/version info from the hyperfleet-api CLI.
- Condition subfield queries for selective Sentinel polling **PR:** [#71](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/71) **Jira:** [HYPERFLEET-536](https://issues.redhat.com/browse/HYPERFLEET-536)
  - Lets Sentinel query narrower slices of condition data instead of full status payloads.
- Slice field validation in SliceFilter **PR:** [#78](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/78), [#69](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/69) **Jira:** [HYPERFLEET-705](https://issues.redhat.com/browse/HYPERFLEET-705), [HYPERFLEET-694](https://issues.redhat.com/browse/HYPERFLEET-694)
  - Validates presenter slice filters so malformed slice expressions fail fast; complements broader resilience work under cascading-failure hardening.
- Connection retry logic for database sidecar startup coordination **PR:** [#69](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/69) **Jira:** [HYPERFLEET-694](https://issues.redhat.com/browse/HYPERFLEET-694)
  - Retries DB connections when the app and sidecar come up out of order.
- pgbouncer sidecar to Helm chart for connection pooling **PR:** [#69](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/69) **Jira:** [HYPERFLEET-694](https://issues.redhat.com/browse/HYPERFLEET-694)
  - Adds optional PgBouncer alongside the API for pooled PostgreSQL access.
- Health check ping timeout configuration **PR:** [#69](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/69) **Jira:** [HYPERFLEET-694](https://issues.redhat.com/browse/HYPERFLEET-694)
  - Makes DB health ping timeouts configurable to match deployment probes.
- Request-level context timeout to transaction middleware **PR:** [#69](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/69) **Jira:** [HYPERFLEET-694](https://issues.redhat.com/browse/HYPERFLEET-694)
  - Bounds request handling time at the transaction middleware layer.
- Connection pool timeout configuration **PR:** [#69](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/69) **Jira:** [HYPERFLEET-694](https://issues.redhat.com/browse/HYPERFLEET-694)
  - Exposes pool wait/timeout settings for database clients.
- PostgreSQL advisory locks for migration coordination in multi-replica deployments **PR:** [#72](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/72) **Jira:** [HYPERFLEET-618](https://issues.redhat.com/browse/HYPERFLEET-618)
  - Serializes migrations across replicas using advisory locks.
- Search and filtering documentation **PR:** [#63](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/63) **Jira:** [HYPERFLEET-299](https://issues.redhat.com/browse/HYPERFLEET-299)
  - Documents the API search query language and filtering behavior.
- Connection pool and pgbouncer documentation **PR:** [#69](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/69) **Jira:** [HYPERFLEET-694](https://issues.redhat.com/browse/HYPERFLEET-694)
  - Documents pooling and PgBouncer deployment options.
- HyperFleet API operator guide **PR:** [#76](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/76) **Jira:** [HYPERFLEET-634](https://issues.redhat.com/browse/HYPERFLEET-634)
  - Operator-focused guide for running and configuring the API.
- OpenTelemetry tracing configuration and instrumentation **PR:** [#89](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/89) **Jira:** [HYPERFLEET-762](https://issues.redhat.com/browse/HYPERFLEET-762)
  - Adds configurable OTLP tracing and related dependency updates (including gRPC CVE fix).
- Contributing guide aligned with architecture standard **PR:** [#95](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/95) **Jira:** [HYPERFLEET-760](https://issues.redhat.com/browse/HYPERFLEET-760)
  - Adds `CONTRIBUTING.md` and versioned changelog practices per platform documentation standards.
- Expanded unit test coverage for errors package **PR:** [#93](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/93) **Jira:** [HYPERFLEET-298](https://issues.redhat.com/browse/HYPERFLEET-298)
  - Broadens tests around service error helpers and edge cases.

### Changed

- BREAKING CHANGE: Cluster and NodePool resource IDs now use RFC 4122 UUID v7 **PR:** [#98](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/98) **Jira:** [HYPERFLEET-732](https://issues.redhat.com/browse/HYPERFLEET-732)
  - Aligns API identifiers with Hypershift `spec.clusterID` expectations using time-ordered UUIDs.
- Standardized appVersion and image.tag handling in Helm chart **PR:** [#90](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/90) **Jira:** [HYPERFLEET-794](https://issues.redhat.com/browse/HYPERFLEET-794)
  - Makes chart app version and image tag conventions consistent with other HyperFleet charts.
- Aligned Helm chart with conventions standard **PR:** [#87](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/87) **Jira:** [HYPERFLEET-786](https://issues.redhat.com/browse/HYPERFLEET-786)
  - Brings the API chart in line with the shared Helm standard.
- Streamlined configuration system with Viper, removed getters and `_FILE` suffix pattern **PR:** [#75](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/75) **Jira:** [HYPERFLEET-685](https://issues.redhat.com/browse/HYPERFLEET-685)
  - Centralizes config loading via Viper and drops the older accessor/`_FILE` env pattern.
- Used `CHANGE_ME` placeholder for image registry **PR:** [#83](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/83) **Jira:** [HYPERFLEET-646](https://issues.redhat.com/browse/HYPERFLEET-646)
  - Avoids misleading default registry values; operators must set an explicit registry.
- Read-only requests skip write transactions; DAO uses request-scoped DB sessions **PR:** [#96](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/96) **Jira:** [HYPERFLEET-724](https://issues.redhat.com/browse/HYPERFLEET-724)
  - Reduces unnecessary transactions on GETs and ensures writes participate in the middleware transaction.

### Fixed

- Validated adapter status conditions in handler layer **PR:** [#88](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/88) **Jira:** [HYPERFLEET-647](https://issues.redhat.com/browse/HYPERFLEET-647)
  - Rejects adapter status payloads with empty or invalid condition entries before service logic runs.
- Removed org prefix from image.repository default **PR:** [#86](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/86) **Jira:** [HYPERFLEET-785](https://issues.redhat.com/browse/HYPERFLEET-785)
  - Matches API chart defaults to adapter/sentinel repository naming.
- Addressed revive linter violations from enabled linting standard **PR:** [#85](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/85) **Jira:** [HYPERFLEET-588](https://issues.redhat.com/browse/HYPERFLEET-588)
  - Clears revive findings so the stricter lint profile stays green.
- Truncated migrations table in CleanDB to ensure migrations re-run **PR:** [#72](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/72) **Jira:** [HYPERFLEET-618](https://issues.redhat.com/browse/HYPERFLEET-618)
  - Test cleanup resets migration state so suites can replay migrations reliably.
- Added fallback default for AdvisoryLockTimeout **PR:** [#72](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/72) **Jira:** [HYPERFLEET-618](https://issues.redhat.com/browse/HYPERFLEET-618)
  - Prevents zero/empty timeout from breaking advisory lock acquisition behavior.
- Rejected `not` operator for condition queries **PR:** [#80](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/80) **Jira:** [HYPERFLEET-709](https://issues.redhat.com/browse/HYPERFLEET-709)
  - Fails unsupported negation queries instead of returning misleading empty results.
- SliceFilter star propagation in query processing **PR:** [#79](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/79) **Jira:** [HYPERFLEET-705](https://issues.redhat.com/browse/HYPERFLEET-705)
  - Corrects wildcard propagation so slice filters match intended fields.
- Used `0.0.0-dev` version for dev image builds **PR:** [#77](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/77) **Jira:** [HYPERFLEET-734](https://issues.redhat.com/browse/HYPERFLEET-734)
  - Dev images no longer inherit arbitrary git tag versions in the binary metadata.
- Aligned health ping timeout with K8s probe timeout **PR:** [#69](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/69) **Jira:** [HYPERFLEET-694](https://issues.redhat.com/browse/HYPERFLEET-694)
  - Keeps DB readiness checks within probe deadline budgets.
- Hardened pgbouncer config and health check responses **PR:** [#69](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/69) **Jira:** [HYPERFLEET-694](https://issues.redhat.com/browse/HYPERFLEET-694)
  - Tightens PgBouncer settings and health endpoint behavior for sidecar operation.
- pgbouncer secret handling, connection leak, and lint **PR:** [#69](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/69) **Jira:** [HYPERFLEET-694](https://issues.redhat.com/browse/HYPERFLEET-694)
  - Fixes credential wiring, leaks, and lint issues in the pooling path.
- Return HTTP 503 when the database is unreachable **PR:** [#97](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/97) **Jira:** [HYPERFLEET-659](https://issues.redhat.com/browse/HYPERFLEET-659)
  - Surfaces service-unavailable instead of internal error when PostgreSQL is down.
- Prevent advisory lock race when a transaction starts before the lock **PR:** [#101](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/101) **Jira:** [HYPERFLEET-875](https://issues.redhat.com/browse/HYPERFLEET-875)
  - Orders lock acquisition relative to transactions to avoid migration deadlocks/races.
- Removed hardcoded JWT token from log masking test **PR:** [#100](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/100) **Jira:** [HYPERFLEET-803](https://issues.redhat.com/browse/HYPERFLEET-803)
  - Eliminates embedded credential material from the test suite.

## [0.1.1](https://github.com/openshift-hyperfleet/hyperfleet-api/compare/v0.1.0...v0.1.1) - 2026-03-09

### Added

- Test suite for presenter package **PR:** [#64](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/64) **Jira:** [HYPERFLEET-301](https://issues.redhat.com/browse/HYPERFLEET-301)
  - Adds unit tests covering presenter formatting and parsing paths.
- DatabaseConfig test coverage and improved advisory lock tests **PR:** [#72](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/72) **Jira:** [HYPERFLEET-618](https://issues.redhat.com/browse/HYPERFLEET-618)
  - Extends tests around DB config and concurrent migration locking (same epic as advisory locks).
- Prometheus metrics with `hyperfleet_db_` prefix to database layer **PR:** [#58](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/58) **Jira:** [HYPERFLEET-470](https://issues.redhat.com/browse/HYPERFLEET-470)
  - Exposes database-layer metrics with a consistent `hyperfleet_db_` prefix.

### Changed

- Updated copyright year to 2026 **PR:** [#58](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/58) **Jira:** [HYPERFLEET-470](https://issues.redhat.com/browse/HYPERFLEET-470)
  - Routine notice file update shipped with the metrics change set.
- Renamed metrics to use `hyperfleet_api_` prefix for consistency **PR:** [#57](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/57) **Jira:** [HYPERFLEET-457](https://issues.redhat.com/browse/HYPERFLEET-457)
  - Normalizes application metric names for Prometheus scraping.
- Standardized Dockerfiles and Makefile for building images **PR:** [#59](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/59) **Jira:** [HYPERFLEET-509](https://issues.redhat.com/browse/HYPERFLEET-509)
  - Aligns container build entrypoints and Make targets across environments.

### Fixed

- CA certificates missing in ubi9-micro runtime image **PR:** [#74](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/74) **Jira:** [HYPERFLEET-730](https://issues.redhat.com/browse/HYPERFLEET-730)
  - Restores TLS trust roots so outbound HTTPS from the micro image succeeds.
- VERSION collision with go-toolset base image **PR:** [#70](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/70) **Jira:** [HYPERFLEET-723](https://issues.redhat.com/browse/HYPERFLEET-723)
  - Stops the base image `VERSION` env from corrupting application version reporting.
- Config file resolution broken by `-trimpath` build flag **PR:** [#66](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/66) **Jira:** [HYPERFLEET-710](https://issues.redhat.com/browse/HYPERFLEET-710)
  - Ensures embedded or relative config paths resolve in trimmed binaries.
- Enforced mandatory conditions in adapter status **PR:** [#60](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/60) **Jira:** [HYPERFLEET-657](https://issues.redhat.com/browse/HYPERFLEET-657)
  - Prevents partial adapter reports from overwriting required cluster conditions incorrectly.
- SliceFilter usage in handlers and time field handling **PR:** [#64](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/64) **Jira:** [HYPERFLEET-301](https://issues.redhat.com/browse/HYPERFLEET-301)
  - Corrects handler integration with SliceFilter including time fields.
- Helm chart testing and default image registry **PR:** [#62](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/62) **Jira:** [HYPERFLEET-658](https://issues.redhat.com/browse/HYPERFLEET-658)
  - Adds chart test coverage and consistent registry defaults like other components.
- Reset and re-seed buildInfoMetric in ResetMetricCollectors **PR:** [#57](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/57) **Jira:** [HYPERFLEET-457](https://issues.redhat.com/browse/HYPERFLEET-457)
  - Lets tests reset build info metrics without stale labels.
- Rejected creation requests with missing spec field **PR:** [#56](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/56) **Jira:** [HYPERFLEET-649](https://issues.redhat.com/browse/HYPERFLEET-649)
  - Rejects null or omitted `spec` on create per OpenAPI rules.

## [0.1.0](https://github.com/openshift-hyperfleet/hyperfleet-api/compare/c33867f...v0.1.0) - 2026-02-16

### Added

- PodDisruptionBudget to Helm chart **PR:** [#44](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/44) **Jira:** [HYPERFLEET-579](https://issues.redhat.com/browse/HYPERFLEET-579)
  - Protects API availability during voluntary disruptions.
- ServiceMonitor to Helm chart for Prometheus Operator integration **PR:** [#43](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/43) **Jira:** [HYPERFLEET-581](https://issues.redhat.com/browse/HYPERFLEET-581)
  - Enables Prometheus Operator to scrape API metrics.
- YAML table format for adapter requirements **PR:** [#41](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/41) **Jira:** [HYPERFLEET-29](https://issues.redhat.com/browse/HYPERFLEET-29)
  - Documents required adapters in a structured YAML table.
- Configurable adapter requirements **PR:** [#40](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/40) **Jira:** [HYPERFLEET-29](https://issues.redhat.com/browse/HYPERFLEET-29)
  - Makes which adapters are required a configuration concern.
- Condition-based search with GIN index for improved query performance **PR:** [#39](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/39) **Jira:** [HYPERFLEET-386](https://issues.redhat.com/browse/HYPERFLEET-386)
  - Adds indexed queries over condition data for faster searches.
- Health endpoints (`/healthz`, `/readyz`) and graceful shutdown **PR:** [#34](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/34) **Jira:** [HYPERFLEET-453](https://issues.redhat.com/browse/HYPERFLEET-453), [HYPERFLEET-454](https://issues.redhat.com/browse/HYPERFLEET-454)
  - Standard probes plus signal-aware shutdown (split across health and lifecycle tickets).
- User-friendly search syntax with lowercase Base32 ID encoding **PR:** [#16](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/16) **Jira:** [HYPERFLEET-321](https://issues.redhat.com/browse/HYPERFLEET-321)
  - Improves search UX and stable encoding of identifiers in queries.
- Schema validation for cluster and nodepool specifications **PR:** [#12](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/12) **Jira:** *(not linked in merge commits)*
  - Validates submitted specs against the OpenAPI schema at the edge.
- Generation field for NodePool management **PR:** [#22](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/22) **Jira:** *(not linked in merge commits)*
  - Introduces generation for optimistic concurrency on node pool updates.
- OpenAPI schema embedded in Docker image for runtime validation **PR:** [#14](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/14) **Jira:** [HYPERFLEET-312](https://issues.redhat.com/browse/HYPERFLEET-312)
  - Ships the spec inside the image to support runtime validation targets.
- Helm chart for Kubernetes deployment **PR:** [#16](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/16) **Jira:** [HYPERFLEET-321](https://issues.redhat.com/browse/HYPERFLEET-321)
  - Delivers a deployable chart alongside search work in the same change set.
- gomock/mockgen infrastructure for service mocks **PR:** [#10](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/10) **Jira:** [HYPERFLEET-290](https://issues.redhat.com/browse/HYPERFLEET-290)
  - Standardizes generated mocks for services and tests.
- Bingo for Go tool dependency management **PR:** [#9](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/9) **Jira:** [HYPERFLEET-27](https://issues.redhat.com/browse/HYPERFLEET-27)
  - Pins toolchain versions (lint, codegen, etc.) via Bingo.
- Linux/amd64 platform support for container builds **PR:** [#17](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/17) **Jira:** *(not linked in merge commits)*
  - Constrains image builds to linux/amd64 for consistent artifacts.
- Integration tests for conditions **PR:** [#39](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/39) **Jira:** [HYPERFLEET-386](https://issues.redhat.com/browse/HYPERFLEET-386)
  - End-to-end coverage for the conditions-based status model.
- Dynamic table discovery for test cleanup **PR:** [#32](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/32) **Jira:** [HYPERFLEET-302](https://issues.redhat.com/browse/HYPERFLEET-302)
  - Integration tests discover tables dynamically when resetting state.
- Operational runbook and metrics documentation **PR:** [#45](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/45) **Jira:** [HYPERFLEET-580](https://issues.redhat.com/browse/HYPERFLEET-580)
  - Documents day-two operations and observability for the API.
- ServiceMonitor configuration documentation **PR:** [#43](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/43) **Jira:** [HYPERFLEET-581](https://issues.redhat.com/browse/HYPERFLEET-581)
  - Explains how to configure scraping via ServiceMonitor.
- Params constraints documentation **PR:** [#55](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/55) **Jira:** [HYPERFLEET-631](https://issues.redhat.com/browse/HYPERFLEET-631)
  - Documents naming constraints for clusters and node pools.

### Changed

- BREAKING CHANGE: Updated OpenAPI spec for conditions-based status model **PR:** [#39](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/39) **Jira:** [HYPERFLEET-386](https://issues.redhat.com/browse/HYPERFLEET-386)
  - Moves status to Kubernetes-style conditions and updates the public schema accordingly.
- Aligned cluster and nodepool name validation with CS rules **PR:** [#48](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/48) **Jira:** [HYPERFLEET-364](https://issues.redhat.com/browse/HYPERFLEET-364)
  - Matches name validation to cluster service naming rules.
- Implemented RFC 9457 Problem Details error model for standardized error responses **PR:** [#37](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/37) **Jira:** [HYPERFLEET-452](https://issues.redhat.com/browse/HYPERFLEET-452)
  - Returns structured problem+json compatible errors across the API.
- Migrated to oapi-codegen for OpenAPI code generation **PR:** [#33](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/33) **Jira:** [HYPERFLEET-501](https://issues.redhat.com/browse/HYPERFLEET-501)
  - Replaces openapi-generator-cli Docker flows with native oapi-codegen and Bingo-managed tooling.
- Aligned logging with HyperFleet structured logging specification **PR:** [#31](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/31) **Jira:** [HYPERFLEET-363](https://issues.redhat.com/browse/HYPERFLEET-363)
  - Adopts shared structured logging conventions for the API component.
- Integrated database logging with LOG_LEVEL **PR:** [#35](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/35) **Jira:** *(not linked in merge commits)*
  - Ties database log verbosity to the global log level configuration.
- Renamed Makefile binary target to build with output to bin/ **PR:** [#30](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/30) **Jira:** [HYPERFLEET-456](https://issues.redhat.com/browse/HYPERFLEET-456)
  - Standardizes build output location for the hyperfleet-api binary.
- Consolidated and streamlined documentation structure **PR:** [#21](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/21) **Jira:** *(not linked in merge commits)*
  - Reorganizes docs for the MVP API surface.
- Configured rh-hooks-ai for AI-readiness and security compliance **PR:** [#18](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/18) **Jira:** *(not linked in merge commits)*
  - Adds Red Hat AI/security pre-commit hooks to the repo.
- Migrated to HyperFleet v2 architecture **PR:** [#3](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/3) **Jira:** [HYPERFLEET-3](https://issues.redhat.com/browse/HYPERFLEET-3)
  - Foundational service layout and module rename for the v2 API codebase.

### Removed

- Phase validation from status types in favor of conditions-based model **PR:** [#39](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/39) **Jira:** [HYPERFLEET-386](https://issues.redhat.com/browse/HYPERFLEET-386)
  - Drops legacy phase validation as conditions become authoritative.
- Generated mock files from git tracking **PR:** [#10](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/10) **Jira:** [HYPERFLEET-290](https://issues.redhat.com/browse/HYPERFLEET-290)
  - Mocks are generated locally instead of committed.
- Generated OpenAPI code from git tracking **PR:** [#3](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/3) **Jira:** [HYPERFLEET-3](https://issues.redhat.com/browse/HYPERFLEET-3)
  - Generated client/models are produced at build time rather than stored in git.
- .claude directory from git tracking **PR:** [#45](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/45) **Jira:** [HYPERFLEET-580](https://issues.redhat.com/browse/HYPERFLEET-580)
  - Stops committing editor-specific assistant metadata with the runbook work.

### Fixed

- Prevented duplicate nodepool names within a cluster **PR:** [#53](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/53) **Jira:** [HYPERFLEET-619](https://issues.redhat.com/browse/HYPERFLEET-619)
  - Enforces unique node pool names per cluster on create.
- Returned 404 for non-existent cluster statuses **PR:** [#54](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/54) **Jira:** [HYPERFLEET-615](https://issues.redhat.com/browse/HYPERFLEET-615)
  - Status GETs for missing clusters now return not found instead of empty success.
- First adapter status report now correctly initializes with Available=Unknown **PR:** [#52](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/52) **Jira:** [HYPERFLEET-608](https://issues.redhat.com/browse/HYPERFLEET-608)
  - Matches Kubernetes semantics for initial condition state before evidence arrives.
- Integration tests updated to match new first-report behavior **PR:** [#52](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/52) **Jira:** [HYPERFLEET-608](https://issues.redhat.com/browse/HYPERFLEET-608)
  - Aligns tests with the corrected first-report condition defaults.
- Added timeout to testcontainer teardown to prevent Prow hang **PR:** [#52](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/52) **Jira:** [HYPERFLEET-608](https://issues.redhat.com/browse/HYPERFLEET-608)
  - Avoids CI hangs during container shutdown in integration tests.
- Centralized adapter config to avoid duplicate logs **PR:** [#46](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/46) **Jira:** [HYPERFLEET-606](https://issues.redhat.com/browse/HYPERFLEET-606)
  - Loads adapter requirements once to reduce noisy duplicate logging.
- Avoided exposing secret values in runbook **PR:** [#45](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/45) **Jira:** [HYPERFLEET-580](https://issues.redhat.com/browse/HYPERFLEET-580)
  - Redacts sensitive examples from operational documentation.
- Made adapter configuration mandatory **PR:** [#46](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/46) **Jira:** [HYPERFLEET-606](https://issues.redhat.com/browse/HYPERFLEET-606)
  - Requires explicit adapter configuration at startup.
- Used explicit nil checks for PDB values **PR:** [#44](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/44) **Jira:** [HYPERFLEET-579](https://issues.redhat.com/browse/HYPERFLEET-579)
  - Avoids rendering invalid PDB fragments when values are unset.
- Fixed goconst, gocritic, gosec, unparam and lll lint issues **PR:** [#42](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/42) **Jira:** [HYPERFLEET-455](https://issues.redhat.com/browse/HYPERFLEET-455)
  - Brings the repo in line with the expanded golangci-lint configuration.
- Prevented fmt.Sprintf panic when reason contains `%` without values **PR:** [#37](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/37) **Jira:** [HYPERFLEET-452](https://issues.redhat.com/browse/HYPERFLEET-452)
  - Hardens problem detail formatting against printf-style strings in reasons.
- Avoided leaking database error details to API clients **PR:** [#37](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/37) **Jira:** [HYPERFLEET-452](https://issues.redhat.com/browse/HYPERFLEET-452)
  - Maps internal DB errors to safe client-facing messages.
- Omitted empty Instance and TraceId from Problem Details JSON **PR:** [#37](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/37) **Jira:** [HYPERFLEET-452](https://issues.redhat.com/browse/HYPERFLEET-452)
  - Keeps problem responses minimal when optional fields are unused.
- Added missing error codes to errorDefinitions map **PR:** [#37](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/37) **Jira:** [HYPERFLEET-452](https://issues.redhat.com/browse/HYPERFLEET-452)
  - Ensures every service error maps to a documented code in problem responses.
- MVP phase logic to only return Ready or NotReady **PR:** [#9](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/9) **Jira:** [HYPERFLEET-27](https://issues.redhat.com/browse/HYPERFLEET-27)
  - Simplifies early lifecycle phase reporting for adapter status.
- Cluster and nodepool name validation **PR:** [#16](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/16) **Jira:** [HYPERFLEET-321](https://issues.redhat.com/browse/HYPERFLEET-321)
  - Validates names on create/update as part of search and API hardening.
- Silent error suppression **PR:** [#26](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/26) **Jira:** [HYPERFLEET-297](https://issues.redhat.com/browse/HYPERFLEET-297)
  - Surfaces handler errors instead of swallowing them.
- Propagated JSON unmarshal errors **PR:** [#26](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/26) **Jira:** [HYPERFLEET-297](https://issues.redhat.com/browse/HYPERFLEET-297)
  - Returns parse failures to clients with appropriate problem details.
- Lint failures in presubmit jobs **PR:** [#12](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/12), [#6](https://github.com/openshift-hyperfleet/hyperfleet-api/pull/6) **Jira:** *(not linked in merge commits)*
  - Fixes CI lint/unit failures blocking merges for early schema work.

[0.2.0]: https://github.com/openshift-hyperfleet/hyperfleet-api/compare/v0.1.1...v0.2.0
