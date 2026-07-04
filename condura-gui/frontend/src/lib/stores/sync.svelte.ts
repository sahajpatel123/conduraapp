// Sync store (Phase 14F).
//
// Replaces the inline pairing logic currently in Sync.svelte
// (and any other component that previously called the sync.*
// RPCs directly). The store caches the discovered peers and
// paired devices, and exposes the high-level pairing/revocation
// methods the GUI needs.
//
// State machine:
//   - The daemon is the source of truth for peer discovery
//     and the paired set. We re-sync on mount and after every
//     mutation.
//   - `pendingPin` is ephemeral: the user reads the PIN on
//     the existing device, types it on the new device. The
//     PIN is NOT persisted.
//
//   ┌──────────────┐  pairWith(peer)    ┌──────────────┐
//   │   no peers   │ ──────────────────▶ │  pendingPin  │
//   │              │                     │              │
//   └──────────────┘                     └──────────────┘
//        ▲                                       │
//        │                                       │
//        │ refresh()  ◀────  confirm() / cancel  │
//
// The daemon's sync.pair_begin returns a 5-min TTL. We don't
// surface that here because the GUI prompts the user for the
// PIN within seconds, but it's worth knowing the underlying
// state does expire.

import { ipc } from '../ipc/client'
import type {
  SyncPeer,
  SyncPair,
  PairBeginResult,
  PairConfirmResult,
} from '../ipc/types'

/**
 * SyncStore: pairing + paired-set state. Methods are
 * thin wrappers around the daemon's sync.* RPCs that
 * refresh local state on success.
 *
 * Lifecycle:
 *   1. App mounts → store starts with empty peers/pairs.
 *   2. Sync.svelte calls sync.refresh() to fetch.
 *   3. The user clicks "Pair" on a peer → pairWith(peer.id).
 *   4. The store caches the pendingPin and the expiresAt.
 *   5. The user types the PIN on the new device (or pastes
 *      it from the existing device) and clicks "Confirm" →
 *      confirmPairing(pin).
 *   6. The daemon validates the PIN against the stored
 *      token; on match, the new device is added to the
 *      paired set.
 *   7. The store re-syncs to pick up the new paired device.
 */
export class SyncStore {
  /** Discovered peers on the LAN. */
  peers = $state<SyncPeer[]>([])

  /** Paired (trusted) devices. */
  pairs = $state<SyncPair[]>([])

  /** True while a refresh / pair / revoke RPC is in flight. */
  loading = $state<boolean>(false)

  /**
   * The most recent error from a sync RPC. Cleared on the
   * next call. Surfaced as an inline error message.
   */
  error = $state<string | null>(null)

  /**
   * The PIN awaiting confirmation. Empty when no pairing
   * is in flight. Ephemeral — not persisted.
   */
  pendingPin = $state<string>('')

  /**
   * The peer ID awaiting confirmation. Together with
   * pendingPin this represents the in-flight pairing.
   */
  pendingPeerId = $state<string>('')

  /**
   * The expires_at timestamp (RFC 3339) of the pending
   * pairing, returned by sync.pair_begin. The GUI shows
   * a countdown.
   */
  pendingExpiresAt = $state<string>('')

  /**
   * Fetches both the peer list and the paired-set from
   * the daemon. Called on mount and after every mutation.
   */
  async refresh(): Promise<void> {
    this.loading = true
    this.error = null
    try {
      const [peersRes, pairsRes] = await Promise.all([
        ipc.syncPeers(),
        ipc.syncListPairs(),
      ])
      this.peers = peersRes.peers ?? []
      this.pairs = pairsRes.devices ?? []
    } catch (e) {
      this.error = String(e)
    } finally {
      this.loading = false
    }
  }

  /**
   * Starts pairing with the given peer. The daemon mints a
   * one-time token, returns a 6-digit PIN, and stores the
   * token in pendingPairings. We cache the PIN so the GUI
   * can display it; the user reads it on the new device
   * and types it on the existing device to confirm.
   */
  async pairWith(peerId: string): Promise<PairBeginResult | null> {
    this.loading = true
    this.error = null
    try {
      const result = await ipc.syncPairBeginTyped(peerId)
      this.pendingPeerId = peerId
      this.pendingPin = result.pin
      // expires_in is seconds; convert to ISO timestamp
      // for the GUI's countdown display.
      const expiresAt = new Date(Date.now() + result.expires_in * 1000)
      this.pendingExpiresAt = expiresAt.toISOString()
      return result
    } catch (e) {
      this.error = String(e)
      this.clearPending()
      return null
    } finally {
      this.loading = false
    }
  }

  /**
   * Confirms a pending pairing. The user has read the PIN
   * from the new device and typed it on the existing
   * device. The daemon validates the PIN against the
   * stored token; on match, the new device is added to
   * the paired set.
   */
  async confirmPairing(pin: string): Promise<PairConfirmResult | null> {
    if (!this.pendingPeerId) {
      this.error = 'No pending pairing to confirm'
      return null
    }
    this.loading = true
    this.error = null
    try {
      const result = await ipc.syncPairConfirmTyped(this.pendingPeerId, pin)
      this.clearPending()
      await this.refresh()
      return result
    } catch (e) {
      this.error = String(e)
      return null
    } finally {
      this.loading = false
    }
  }

  /**
   * Cancels a pending pairing. The daemon's pending token
   * expires on its own TTL; this just clears our local
   * state so the GUI returns to the peer list.
   */
  clearPending(): void {
    this.pendingPeerId = ''
    this.pendingPin = ''
    this.pendingExpiresAt = ''
  }

  /**
   * Revokes a paired device. The daemon removes the
   * device from the paired set and signs a revocation
   * message (for propagation to other paired devices).
   */
  async revoke(deviceId: string): Promise<boolean> {
    this.loading = true
    this.error = null
    try {
      await ipc.syncRevokeTyped(deviceId)
      await this.refresh()
      return true
    } catch (e) {
      this.error = String(e)
      return false
    } finally {
      this.loading = false
    }
  }

  /**
   * Returns the peer with the given ID, or null if not
   * found. Used by the GUI to look up peer metadata
   * (name, fingerprint) when displaying the pending
   * pairing screen.
   */
  peerById(peerId: string): SyncPeer | null {
    return this.peers.find((p) => p.device_id === peerId) ?? null
  }
}

// Singleton instance — only one pairing flow at a time.
export const sync = new SyncStore()
