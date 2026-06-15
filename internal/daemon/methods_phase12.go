package daemon

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sahajpatel123/synapticapp/internal/hub"
	"github.com/sahajpatel123/synapticapp/internal/i18n"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
	"github.com/sahajpatel123/synapticapp/internal/skills"
	"github.com/sahajpatel123/synapticapp/internal/sync"
)

// Phase12Components bundles the Phase 12 subsystems.
type Phase12Components struct {
	HubClient  *hub.Client
	SkillStore *skills.SQLiteStore
	SyncEngine *sync.Engine
	Catalog    *i18n.Catalog
}

// registerPhase12Methods wires hub.*, sync.*, i18n.*, and skills.* RPC methods.
func registerPhase12Methods(srv *ipc.Server, p12 *Phase12Components) {
	if p12 == nil {
		return
	}
	registerHubMethods(srv, p12)
	registerSyncMethods(srv, p12)
	registerI18nMethods(srv, p12)
	registerSkillsMethods(srv, p12)
}

// registerSkillsMethods adds skills.list and skills.get RPC methods.
func registerSkillsMethods(srv *ipc.Server, p12 *Phase12Components) {
	// skills.list: list locally installed skills.
	srv.Register("skills.list", func(ctx context.Context, params json.RawMessage) (any, error) {
		if p12.SkillStore == nil {
			return []*skills.Skill{}, nil
		}
		var p struct {
			Limit int `json:"limit"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p.Limit <= 0 {
			p.Limit = 100
		}
		return p12.SkillStore.List(ctx, p.Limit)
	})

	// skills.get: fetch a single skill by ID.
	srv.Register("skills.get", func(ctx context.Context, params json.RawMessage) (any, error) {
		if p12.SkillStore == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSkillStoreNotAvailable}
		}
		var p struct {
			ID string `json:"id"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p.ID == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "id required"}
		}
		sk, err := p12.SkillStore.Get(ctx, p.ID)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "skill not found"}
		}
		return sk, nil
	})

	// skills.delete: remove a skill by ID.
	srv.Register("skills.delete", func(ctx context.Context, params json.RawMessage) (any, error) {
		if p12.SkillStore == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSkillStoreNotAvailable}
		}
		var p struct {
			ID string `json:"id"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p.ID == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "id required"}
		}
		if err := p12.SkillStore.Delete(ctx, p.ID); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "skill not found"}
		}
		return auditOK(), nil
	})
}

// registerI18nMethods adds i18n.locale and i18n.locales RPC methods.
func registerI18nMethods(srv *ipc.Server, p12 *Phase12Components) {
	// i18n.locales: list available locales.
	srv.Register("i18n.locales", func(_ context.Context, _ json.RawMessage) (any, error) {
		if p12.Catalog == nil {
			return []string{"en"}, nil
		}
		return p12.Catalog.Locales(), nil
	})

	// i18n.locale: return all translations for a locale.
	// Returns raw format strings (with {0} placeholders for the frontend).
	srv.Register("i18n.locale", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Locale string `json:"locale"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p.Locale == "" {
			p.Locale = "en"
		}
		if p12.Catalog == nil {
			return map[string]any{"locale": p.Locale, "translations": map[string]string{}}, nil
		}
		// Return raw format strings from the locale files.
		// The frontend's t() function handles {0} placeholder replacement.
		translations := p12.Catalog.RawTranslations(p.Locale)
		return map[string]any{"locale": p.Locale, "translations": translations}, nil
	})
}

const errHubNotConfigured = "hub not configured"
const errSkillStoreNotAvailable = "skill store not available"

func registerHubMethods(srv *ipc.Server, p12 *Phase12Components) {
	// hub.search: search the Skills Hub for skills.
	srv.Register("hub.search", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Query string `json:"query"`
			Limit int    `json:"limit"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p.Limit <= 0 {
			p.Limit = 20
		}
		if p12.HubClient == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errHubNotConfigured}
		}
		return p12.HubClient.Search(p.Query, p.Limit)
	})

	// hub.get: fetch skill metadata from the hub.
	srv.Register("hub.get", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			ID string `json:"id"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p12.HubClient == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errHubNotConfigured}
		}
		return p12.HubClient.Get(p.ID)
	})

	// hub.install: download, scan, and install a skill from the hub.
	srv.Register("hub.install", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			ID string `json:"id"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p12.HubClient == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errHubNotConfigured}
		}
		if p12.SkillStore == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSkillStoreNotAvailable}
		}
		return installSkillFromHub(ctx, p.ID, p12.HubClient, p12.SkillStore)
	})

	// hub.publish: upload a local skill to the hub.
	srv.Register("hub.publish", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			ID string `json:"id"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p12.HubClient == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errHubNotConfigured}
		}
		if p12.SkillStore == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSkillStoreNotAvailable}
		}
		return publishSkillToHub(ctx, p.ID, p12.SkillStore, p12.HubClient)
	})
}

func installSkillFromHub(ctx context.Context, id string, client *hub.Client, store *skills.SQLiteStore) (any, error) {
	// Download from hub.
	data, checksum, err := client.Download(id)
	if err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
	}

	// Verify checksum.
	if err := hub.Verify(data, checksum); err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "checksum verification failed: " + err.Error()}
	}

	// Safety scan.
	scan := hub.Scan(data)
	if !scan.Safe {
		return nil, &ipc.Error{Code: ipc.CodeInternalError,
			Message: fmt.Sprintf("skill failed safety scan: %v", scan.Issues)}
	}

	// Fetch metadata.
	meta, err := client.Get(id)
	if err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
	}

	// Install into local store from archive bytes.
	parsed, err := skills.ParseArchive(data)
	if err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
	}
	sk := parsed
	if sk.Trust == "" {
		sk.Trust = skills.TrustCommunity
	}
	sk.Source = "hub"
	sk.HubID = meta.ID
	sk.Checksum = checksum
	if sk.Author == "" {
		sk.Author = meta.Author
	}
	if sk.License == "" {
		sk.License = meta.License
	}
	if err := store.Create(ctx, sk); err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "install: " + err.Error()}
	}
	return map[string]any{"ok": true, "id": sk.ID}, nil
}

func publishSkillToHub(ctx context.Context, id string, store *skills.SQLiteStore, client *hub.Client) (any, error) {
	sk, err := store.Get(ctx, id)
	if err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "skill not found"}
	}
	meta := hub.SkillMeta{
		ID:          sk.ID,
		Name:        sk.Name,
		Description: sk.Description,
		Version:     sk.Version,
		Author:      sk.Author,
		License:     sk.License,
		Trust:       string(sk.Trust),
	}
	archive, err := skills.MarshalArchive(sk)
	if err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
	}
	if err := client.Publish(archive, meta); err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
	}
	return map[string]any{"ok": true, "id": sk.ID}, nil
}

func registerSyncMethods(srv *ipc.Server, p12 *Phase12Components) {
	// sync.status: return current sync engine status.
	srv.Register("sync.status", func(_ context.Context, _ json.RawMessage) (any, error) {
		if p12.SyncEngine == nil {
			return map[string]any{"enabled": false}, nil
		}
		return p12.SyncEngine.Status(), nil
	})

	// sync.peers: list discovered peers.
	srv.Register("sync.peers", func(_ context.Context, _ json.RawMessage) (any, error) {
		if p12.SyncEngine == nil {
			return map[string]any{"peers": []any{}}, nil
		}
		return map[string]any{"peers": p12.SyncEngine.DiscoveredPeers()}, nil
	})

	// sync.put: store a key-value pair in the CRDT store.
	srv.Register("sync.put", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Key   string `json:"key"`
			Value []byte `json:"value"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p12.SyncEngine == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "sync not enabled"}
		}
		p12.SyncEngine.Put(p.Key, p.Value)
		return auditOK(), nil
	})

	// sync.get: retrieve a value from the CRDT store.
	srv.Register("sync.get", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Key string `json:"key"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p12.SyncEngine == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "sync not enabled"}
		}
		val := p12.SyncEngine.Get(p.Key)
		return map[string]any{"value": val}, nil
	})

	// sync.start: start the sync engine.
	srv.Register("sync.start", func(_ context.Context, _ json.RawMessage) (any, error) {
		if p12.SyncEngine == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "sync not configured"}
		}
		p12.SyncEngine.Start()
		return auditOK(), nil
	})

	// sync.stop: stop the sync engine.
	srv.Register("sync.stop", func(_ context.Context, _ json.RawMessage) (any, error) {
		if p12.SyncEngine == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "sync not configured"}
		}
		p12.SyncEngine.Stop()
		return auditOK(), nil
	})

	// sync.sync_with: one-shot CRDT sync with a discovered peer.
	srv.Register("sync.sync_with", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			DeviceID string `json:"device_id"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p12.SyncEngine == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "sync not enabled"}
		}
		var target *sync.Peer
		for _, peer := range p12.SyncEngine.DiscoveredPeers() {
			if peer.DeviceID == p.DeviceID {
				target = peer
				break
			}
		}
		if target == nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "peer not found"}
		}
		merged, err := p12.SyncEngine.SyncWith(target)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return map[string]any{"ok": true, "merged": merged}, nil
	})
}
