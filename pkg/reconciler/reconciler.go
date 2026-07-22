package reconciler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/openshift-hyperfleet/hyperfleet-api/pkg/api"
	"github.com/openshift-hyperfleet/hyperfleet-api/pkg/config"
	"github.com/openshift-hyperfleet/hyperfleet-api/pkg/dao"
	"github.com/openshift-hyperfleet/hyperfleet-api/pkg/db"
	"github.com/openshift-hyperfleet/hyperfleet-api/pkg/logger"
	"github.com/openshift-hyperfleet/hyperfleet-api/pkg/registry"
)

// eventData is the payload POSTed to each adapter's /reconcile endpoint.
type eventData struct {
	ID              string          `json:"id"`
	Kind            string          `json:"kind"`
	Href            string          `json:"href"`
	Generation      int64           `json:"generation"`
	OwnerReferences *ownerReference `json:"owner_references,omitempty"`
}

type ownerReference struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Href string `json:"href"`
}

// Reconciler polls the database and fires HTTP requests to adapters when resources need reconciliation.
type Reconciler struct {
	cfg            *config.ReconcilerConfig
	sessionFactory db.SessionFactory
	httpClient     *http.Client
}

func New(cfg *config.ReconcilerConfig, sessionFactory db.SessionFactory) *Reconciler {
	return &Reconciler{
		cfg:            cfg,
		sessionFactory: sessionFactory,
		httpClient:     &http.Client{Timeout: cfg.HTTPTimeout},
	}
}

// Start runs the reconcile loop until ctx is cancelled. Intended to run in a goroutine.
func (r *Reconciler) Start(ctx context.Context) {
	if !r.cfg.Enabled {
		logger.Info(ctx, "reconciler disabled, skipping")
		return
	}

	logger.With(ctx,
		"poll_interval", r.cfg.PollInterval,
		"stale_threshold", r.cfg.StaleThreshold,
	).Info("reconciler starting")

	ticker := time.NewTicker(r.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, "reconciler stopped")
			return
		case <-ticker.C:
			r.tick(ctx)
		}
	}
}

func (r *Reconciler) tick(ctx context.Context) {
	resourceDao := dao.NewResourceDao(r.sessionFactory)

	for _, desc := range registry.All() {
		if len(desc.RequiredAdapters) == 0 {
			continue
		}
		r.reconcileKind(ctx, resourceDao, desc)
	}
}

func (r *Reconciler) reconcileKind(ctx context.Context, resourceDao dao.ResourceDao, desc registry.EntityDescriptor) {
	resources, err := resourceDao.FindByKind(ctx, desc.Kind)
	if err != nil {
		logger.With(ctx, "kind", desc.Kind).WithError(err).Error("reconciler: failed to list resources")
		return
	}

	for _, resource := range resources {
		if r.needsReconciliation(resource) {
			r.triggerAdapters(ctx, resource, desc.RequiredAdapters)
		}
	}
}

func (r *Reconciler) needsReconciliation(resource *api.Resource) bool {
	for _, cond := range resource.Conditions {
		if cond.Type == api.ResourceConditionTypeReconciled {
			if cond.Status == api.ConditionFalse {
				return true
			}
			// Reconciled=True but stale — re-trigger
			if time.Since(cond.LastUpdatedTime) > r.cfg.StaleThreshold {
				return true
			}
			return false
		}
	}
	// No Reconciled condition present — treat as needing reconciliation
	return true
}

func (r *Reconciler) triggerAdapters(ctx context.Context, resource *api.Resource, adapters map[string]string) {
	payload, err := buildPayload(resource)
	if err != nil {
		logger.With(ctx, "resource_id", resource.ID).WithError(err).Error("reconciler: failed to build payload")
		return
	}

	for name, url := range adapters {
		go r.postToAdapter(ctx, name, url, payload)
	}
}

func (r *Reconciler) postToAdapter(ctx context.Context, name, url string, payload []byte) {
	endpoint := fmt.Sprintf("%s/reconcile", url)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		logger.With(ctx, "adapter", name, "url", endpoint).WithError(err).Error("reconciler: failed to build request")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		logger.With(ctx, "adapter", name, "url", endpoint).WithError(err).Warn("reconciler: adapter call failed")
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	logger.With(ctx, "adapter", name, "status", resp.StatusCode).Info("reconciler: adapter notified")
}

func buildPayload(resource *api.Resource) ([]byte, error) {
	data := eventData{
		ID:         resource.ID,
		Kind:       resource.Kind,
		Href:       resource.Href,
		Generation: int64(resource.Generation),
	}
	if resource.OwnerID != nil {
		ref := &ownerReference{ID: *resource.OwnerID}
		if resource.OwnerKind != nil {
			ref.Kind = *resource.OwnerKind
		}
		if resource.OwnerHref != nil {
			ref.Href = *resource.OwnerHref
		}
		data.OwnerReferences = ref
	}
	return json.Marshal(data)
}
