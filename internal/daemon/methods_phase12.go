package daemon

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sahajpatel123/conduraapp/internal/config"
	"github.com/sahajpatel123/conduraapp/internal/hub"
	"github.com/sahajpatel123/conduraapp/internal/i18n"
	"github.com/sahajpatel123/conduraapp/internal/ipc"
	"github.com/sahajpatel123/conduraapp/internal/skills"
	"github.com/sahajpatel123/conduraapp/internal/sync"
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
	Config     *config.Config
	Loader     *config.Loader
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
	srv.Register("hub.search", hubSearchHandler(p12))
	srv.Register("hub.get", hubGetHandler(p12))
	srv.Register("hub.install", hubInstallHandler(p12))
	srv.Register("hub.publish", hubPublishHandler(p12))
}

// hubClient returns the configured hub client or an IPC error
// indicating it's not configured.
func hubClient(p12 *Phase12Components) (*hub.Client, error) {
	if p12.HubClient == nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errHubNotConfigured}
	}
	return p12.HubClient, nil
}

func hubSearchHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(_ context.Context, params json.RawMessage) (any, error) {
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
		client, err := hubClient(p12)
		if err != nil {
			return nil, err
		}
		return client.Search(p.Query, p.Limit)
	}
}

func hubGetHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			ID string `json:"id"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		client, err := hubClient(p12)
		if err != nil {
			return nil, err
		}
		return client.Get(p.ID)
	}
}

func hubInstallHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			ID string `json:"id"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		client, err := hubClient(p12)
		if err != nil {
			return nil, err
		}
		if p12.SkillStore == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSkillStoreNotAvail}
		}
		return installSkillFromHub(ctx, p.ID, client, p12.SkillStore)
	}
}

// hub.publish: upload a local skill to the hub. The archive
// bytes are passed in by the caller (the CLI reads the file
// and sends the bytes; the daemon does NOT reach back to disk).
// This way a user can re-publish after editing the archive
// without first importing it into the local store.
func hubPublishHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			ID      string `json:"id"`
			Archive []byte `json:"archive"`
			Path    string `json:"path,omitempty"` // legacy: older CLIs sent a path
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		client, err := hubClient(p12)
		if err != nil {
			return nil, err
		}
		// If archive was not provided, fall back to reading the
		// archive bytes from the local store (legacy behavior).
		if len(p.Archive) == 0 && p.Path != "" && p12.SkillStore != nil {
			sk, gerr := p12.SkillStore.Get(ctx, p.ID)
			if gerr != nil {
				return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "skill not found in local store: " + p.ID}
			}
			p.Archive, gerr = skills.MarshalArchive(sk)
			if gerr != nil {
				return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "marshal archive: " + gerr.Error()}
			}
		}
		if len(p.Archive) == 0 {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "no archive provided (call with --archive <bytes> or --path <local-id>)"}
		}
		return publishSkillToHub(ctx, p.ID, p.Archive, client)
	}
}

func installSkillFromHub(ctx context.Context, id string, client *hub.Client, store *skills.SQLiteStore) (any, error) {
	// Download from hub. Client.Download caps the response size
	// to prevent zip-bomb DoS.
	data, checksum, err := client.Download(id)
	if err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
	}

	// Verify checksum.
	if err := hub.Verify(data, checksum); err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "checksum verification failed: " + err.Error()}
	}

	// Parse the archive BEFORE scanning. hub.Scan operates on the
	// structured skill content (steps, trust, license) — running it
	// on the raw zip bytes produces false positives and rejects
	// every legitimate skill.
	parsed, err := skills.ParseArchive(data)
	if err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "parse archive: " + err.Error()}
	}

	// Safety scan the PARSED skill (not the zip bytes). hub.ScanSkill
	// inspects each step's command for dangerous patterns and
	// verifies trust/license metadata.
	scan := hub.ScanSkill(skillAdapter{sk: parsed})
	if !scan.Safe {
		return nil, &ipc.Error{Code: ipc.CodeInternalError,
			Message: fmt.Sprintf("skill failed safety scan: %v", scan.Issues)}
	}

	// Fetch metadata (used to fill in fields the archive may not
	// carry, like author and license).
	meta, err := client.Get(id)
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

// skillAdapter wraps a *skills.Skill to satisfy hub.SkillScanner
// without forcing the hub package to import skills (which would
// be a cycle).
type skillAdapter struct {
	sk *skills.Skill
}

func (a skillAdapter) ScanSteps() []string { return a.sk.Steps }
func (a skillAdapter) ScanTrust() string   { return string(a.sk.Trust) }
func (a skillAdapter) ScanLicense() string { return a.sk.License }

// publishSkillToHub uploads archive bytes to the hub. Metadata
// is parsed from the archive so the file IS the source of truth
// (no risk of "publish says it shipped but server got stale bytes").
func publishSkillToHub(_ context.Context, id string, archive []byte, client *hub.Client) (any, error) {
	// Parse the archive to extract metadata for the hub. If parsing
	// fails, fall back to bare metadata (just the ID).
	parsed, perr := skills.ParseArchive(archive)
	meta := hub.SkillMeta{ID: id}
	if perr == nil {
		meta = hub.SkillMeta{
			ID:          parsed.ID,
			Name:        parsed.Name,
			Description: parsed.Description,
			Version:     parsed.Version,
			Author:      parsed.Author,
			License:     parsed.License,
			Trust:       string(parsed.Trust),
		}
	} else {
		// Archive is unparseable; publish with just the ID so the
		// server can at least record something.
		meta.Name = id
	}
	if err := client.Publish(archive, meta); err != nil {
		return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
	}
	return map[string]any{"ok": true, "id": id, "bytes": len(archive)}, nil
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
		// Persist so sync state survives daemon restart.
		if p12.Config != nil {
			p12.Config.Sync.Enabled = true
			if p12.Loader != nil {
				_ = p12.Loader.Save(p12.Config)
			}
		}
		return auditOK(), nil
	}
}

func syncStopHandler(p12 *Phase12Components) ipc.HandlerFunc {
	return func(_ context.Context, _ json.RawMessage) (any, error) {
		if p12.SyncEngine == nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: errSyncNotConfigured}
		}
		p12.SyncEngine.Stop()
		if p12.Config != nil {
			p12.Config.Sync.Enabled = false
			if p12.Loader != nil {
				_ = p12.Loader.Save(p12.Config)
			}
		}
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
// The daemon verifies the PIN against the token stored by
// pair_begin and, on match, adds the new device to the paired set.
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
		// The token is stored in the engine from the begin call;
		// the CLI never sees it. This keeps the token off the wire
		// and off the user's screen.
		if err := p12.SyncEngine.ConfirmPairing(target, p.Pin); err != nil {
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
