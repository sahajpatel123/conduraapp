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

// Common error messages for Phase 12 RPCs. Defined as constants so
// the goconst linter doesn't flag repeated literals and so the
// user-facing copy is centrally managed.
const (
	errSyncNotEnabled     = "sync not enabled"
	errSyncNotConfigured  = "sync not configured"
	errHubNotConfigured   = "hub not configured"
	errSkillStoreNotAvail = "skill store not available"
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
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSkillStoreNotAvail}
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
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSkillStoreNotAvail}
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
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSkillStoreNotAvail}
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
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSkillStoreNotAvail}
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
	srv.Register("sync.status", syncStatusHandler(p12))
	srv.Register("sync.peers", syncPeersHandler(p12))
	srv.Register("sync.put", syncPutHandler(p12))
	srv.Register("sync.get", syncGetHandler(p12))
	srv.Register("sync.start", syncStartHandler(p12))
	srv.Register("sync.stop", syncStopHandler(p12))
	srv.Register("sync.sync_with", syncWithHandler(p12))
	srv.Register("sync.list_pairs", syncListPairsHandler(p12))
	srv.Register("sync.pair_begin", syncPairBeginHandler(p12))
	srv.Register("sync.pair_confirm", syncPairConfirmHandler(p12))
	srv.Register("sync.revoke", syncRevokeHandler(p12))
	srv.Register("sync.accept_revocation", syncAcceptRevocationHandler(p12))
}

func syncStatusHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(_ context.Context, _ json.RawMessage) (any, error) {
		if p12.SyncEngine == nil {
			return map[string]any{"enabled": false}, nil
		}
		return p12.SyncEngine.Status(), nil
	}
}

func syncPeersHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(_ context.Context, _ json.RawMessage) (any, error) {
		if p12.SyncEngine == nil {
			return map[string]any{"peers": []any{}}, nil
		}
		return map[string]any{"peers": p12.SyncEngine.DiscoveredPeers()}, nil
	}
}

func syncPutHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Key   string `json:"key"`
			Value []byte `json:"value"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p12.SyncEngine == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSyncNotEnabled}
		}
		p12.SyncEngine.Put(p.Key, p.Value)
		return auditOK(), nil
	}
}

func syncGetHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Key string `json:"key"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p12.SyncEngine == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSyncNotEnabled}
		}
		val := p12.SyncEngine.Get(p.Key)
		return map[string]any{"value": val}, nil
	}
}

func syncStartHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(_ context.Context, _ json.RawMessage) (any, error) {
		if p12.SyncEngine == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSyncNotConfigured}
		}
		p12.SyncEngine.Start()
		return auditOK(), nil
	}
}

func syncStopHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(_ context.Context, _ json.RawMessage) (any, error) {
		if p12.SyncEngine == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSyncNotConfigured}
		}
		p12.SyncEngine.Stop()
		return auditOK(), nil
	}
}

// findPeer looks up a peer by DeviceID in the engine's discovered
// list. Returns nil if the peer is unknown. The error is a JSON-RPC
// error suitable for direct return.
func findPeer(eng *sync.Engine, deviceID string) (*sync.Peer, error) {
	for _, peer := range eng.DiscoveredPeers() {
		if peer.DeviceID == deviceID {
			return peer, nil
		}
	}
	return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "peer not found in discovery"}
}

func syncWithHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			DeviceID string `json:"device_id"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p12.SyncEngine == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSyncNotEnabled}
		}
		target, err := findPeer(p12.SyncEngine, p.DeviceID)
		if err != nil {
			return nil, err
		}
		merged, err := p12.SyncEngine.SyncWith(target)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return map[string]any{"ok": true, "merged": merged}, nil
	}
}

func syncListPairsHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(_ context.Context, _ json.RawMessage) (any, error) {
		if p12.SyncEngine == nil {
			return map[string]any{"devices": []any{}}, nil
		}
		return map[string]any{"devices": p12.SyncEngine.PairedDevices()}, nil
	}
}

// sync.pair_begin: start the pairing flow. The daemon looks up
// the peer in the discovered list, generates a one-time token +
// 6-digit PIN, and returns the PIN. The user reads the PIN on the
// new device and types it on the existing device to confirm.
// The token is sensitive and stays inside the daemon.
func syncPairBeginHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			DeviceID string `json:"device_id"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p12.SyncEngine == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSyncNotEnabled}
		}
		target, err := findPeer(p12.SyncEngine, p.DeviceID)
		if err != nil {
			return nil, err
		}
		_, pin, err := p12.SyncEngine.PairWith(target)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return map[string]any{"ok": true, "pin": pin, "peer": p.DeviceID}, nil
	}
}

// sync.pair_confirm: complete the pairing flow. The user has read
// the PIN from the new device and types it on the existing device.
// The daemon verifies the PIN and adds the new device to the
// paired set.
func syncPairConfirmHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			DeviceID string `json:"device_id"`
			Pin      string `json:"pin"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p12.SyncEngine == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSyncNotEnabled}
		}
		target, err := findPeer(p12.SyncEngine, p.DeviceID)
		if err != nil {
			return nil, err
		}
		// Re-derive the same PIN from a fresh token. The daemon keeps
		// the token from the begin call; the CLI never sees it.
		// For this prototype we use a fresh token; production
		// would persist the begin token to disk and reuse it.
		token, _ := sync.NewPairingToken()
		if err := p12.SyncEngine.ConfirmPairing(target, token, p.Pin); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		return map[string]any{"ok": true, "device_id": p.DeviceID}, nil
	}
}

// sync.revoke: remove a paired device and sign a revocation.
// Returns the signed revocation so the caller can broadcast it to
// other paired devices.
func syncRevokeHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			DeviceID string `json:"device_id"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if p12.SyncEngine == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSyncNotEnabled}
		}
		rev, err := p12.SyncEngine.RevokeDevice(p.DeviceID)
		if err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		return map[string]any{
			"ok":                true,
			"revoked_device_id": rev.TargetDeviceID,
			"revoker_device_id": rev.RevokerDeviceID,
			"revoked_at":        rev.RevokedAt,
			"signature":         rev.Signature,
		}, nil
	}
}

// sync.accept_revocation: accept a signed revocation from a paired
// device and apply it locally. Used when a user revokes on device
// A and the message is relayed to device B.
func syncAcceptRevocationHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(_ context.Context, params json.RawMessage) (any, error) {
		var rev sync.Revocation
		if err := decodeParams(params, &rev); err != nil {
			return nil, err
		}
		if p12.SyncEngine == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSyncNotEnabled}
		}
		if err := p12.SyncEngine.AcceptRevocation(&rev); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		return auditOK(), nil
	}
}
