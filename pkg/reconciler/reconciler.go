package reconciler

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/openshift-hyperfleet/hyperfleet-api/pkg/api"
	"github.com/openshift-hyperfleet/hyperfleet-api/pkg/config"
	"github.com/openshift-hyperfleet/hyperfleet-api/pkg/dao"
	"github.com/openshift-hyperfleet/hyperfleet-api/pkg/db"
	"github.com/openshift-hyperfleet/hyperfleet-api/pkg/logger"
	"github.com/openshift-hyperfleet/hyperfleet-api/pkg/registry"
)

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

type Reconciler struct {
	cfg            *config.ReconcilerConfig
	sessionFactory db.SessionFactory
	directDB       *sql.DB
}

func New(cfg *config.ReconcilerConfig, sessionFactory db.SessionFactory) *Reconciler {
	return &Reconciler{
		cfg:            cfg,
		sessionFactory: sessionFactory,
		directDB:       sessionFactory.DirectDB(),
	}
}

func (r *Reconciler) Start(ctx context.Context) {
	if !r.cfg.Enabled {
		logger.Info(ctx, "reconciler disabled, skipping")
		return
	}

	logger.With(ctx,
		"poll_interval", r.cfg.PollInterval,
		"stale_threshold", r.cfg.StaleThreshold,
	).Info("reconciler starting (db-queue mode)")

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

	notified := false
	for _, desc := range registry.All() {
		if len(desc.RequiredAdapters) == 0 {
			continue
		}
		if r.reconcileKind(ctx, resourceDao, desc) {
			notified = true
		}
	}

	if notified {
		if _, err := r.directDB.ExecContext(ctx, "SELECT pg_notify('messages', '')"); err != nil {
			logger.WithError(ctx, err).Warn("reconciler: pg_notify failed")
		}
	}
}

func (r *Reconciler) reconcileKind(ctx context.Context, resourceDao dao.ResourceDao, desc registry.EntityDescriptor) bool {
	resources, err := resourceDao.FindByKind(ctx, desc.Kind)
	if err != nil {
		logger.With(ctx, "kind", desc.Kind).WithError(err).Error("reconciler: failed to list resources")
		return false
	}

	enqueued := false
	adapterNames := desc.RequiredAdapterNames()
	for _, resource := range resources {
		if r.needsReconciliation(resource) {
			if r.enqueueMessages(ctx, resource, adapterNames) {
				enqueued = true
			}
		}
	}
	return enqueued
}

func (r *Reconciler) needsReconciliation(resource *api.Resource) bool {
	for _, cond := range resource.Conditions {
		if cond.Type == api.ResourceConditionTypeReconciled {
			if cond.Status == api.ConditionFalse {
				return true
			}
			if time.Since(cond.LastUpdatedTime) > r.cfg.StaleThreshold {
				return true
			}
			return false
		}
	}
	return true
}

func (r *Reconciler) enqueueMessages(ctx context.Context, resource *api.Resource, adapterNames []string) bool {
	payload, err := buildPayload(resource)
	if err != nil {
		logger.With(ctx, "resource_id", resource.ID).WithError(err).Error("reconciler: failed to build payload")
		return false
	}

	enqueued := false
	for _, name := range adapterNames {
		result, err := r.directDB.ExecContext(ctx,
			`INSERT INTO messages (adapter_name, kind, resource_id, payload)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT DO NOTHING`,
			name, resource.Kind, resource.ID, payload,
		)
		if err != nil {
			logger.With(ctx, "adapter", name, "resource_id", resource.ID).WithError(err).Warn("reconciler: failed to enqueue message")
			continue
		}
		if rows, _ := result.RowsAffected(); rows > 0 {
			enqueued = true
			logger.With(ctx, "adapter", name, "resource_id", resource.ID).Debug("reconciler: message enqueued")
		}
	}
	return enqueued
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
