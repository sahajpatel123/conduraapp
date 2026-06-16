// Onboarding state (Phase 12/14).
//
// This store is a thin cache of:
//   1. The daemon's authoritative wizard state (via JSON-RPC).
//   2. Ephemeral UI state (EULA doc, error, busy flag) that
//      does NOT belong on the server.
//
// The previous local-only store had `step`, `provider`,
// `apiKey`, `hotkey`, `telemetryEnabled`, `testPassed` fields
// that the daemon never knew about. They were a parallel
// implementation that drifted from the server. This rewrite
// deletes that parallel state and treats the daemon as the
// source of truth for the wizard position. The GUI is a
// thin renderer over the daemon's state machine.
//
// The four-step union (eula → permissions → hotkey → complete)
// is what the user sees. The underlying state machine has
// the same steps now (Phase 14A); legacy 8-step states
// (welcome / power_source / backend_detect / voice_test) are
// transparently migrated on load by the daemon.

import type {
  OnboardingDaemonState,
  EULADocument,
  PowerProbeResult,
  OnboardingFinishParams,
  OnboardingFinishResult,
} from '../ipc/types'
import { ipc } from '../ipc/client'

/**
 * OnboardingStore: thin cache of the daemon's wizard state plus
 * ephemeral UI state. Components subscribe to $state fields
 * and re-render automatically. The daemon is the source of
 * truth; mutations always go through an RPC.
 *
 * Field names are chosen to match the existing Svelte component
 * contract (busy, error, currentStep, daemon, eulaVersion, etc.).
 * Renaming them would require touching every onboarding
 * component.
 */
export class OnboardingStore {
  /** The current daemon state, or null before the first sync. */
  daemon = $state<OnboardingDaemonState | null>(null)

  /** True while a sync or mutation RPC is in flight. */
  busy = $state(false)

  /**
   * Alias for `busy`. The OnboardingWizard.svelte component
   * uses `onboarding.loading`; other components use
   * `onboarding.busy`. Both names refer to the same state.
   */
  get loading(): boolean {
    return this.busy
  }
  set loading(v: boolean) {
    this.busy = v
  }

  /**
   * The EULA document fetched by loadEula(). Cached for the
   * duration of the wizard so the user can navigate back.
   */
  eula = $state<EULADocument | null>(null)

  /**
   * The EULA version the user is about to accept. Set by
   * loadEula() so the GUI can show "you are accepting vN".
   * Used by finish() to record the accepted version.
   */
  eulaVersion = $state<string>('')

  /**
   * The power-probe result, used by the Ready screen to show
   * "found your local Ollama, etc.".
   */
  power = $state<PowerProbeResult | null>(null)

  /**
   * The hotkey the user is currently typing. Ephemeral —
   * not persisted on the server until finish() is called.
   * Components read this and the persisted value (from
   * daemon.steps.hotkey.data) as needed.
   */
  hotkeyValue = $state('')

  /**
   * The most recent error from an RPC. Cleared on the next
   * successful call. The GUI surfaces this as an inline
   * error message.
   */
  error = $state<string | null>(null)

  /**
   * The current step derived from the daemon state. The high-
   * level 4-step union is what the GUI renders. When the
   * daemon reports a legacy step name (welcome, etc.) we
   * fall back to 'eula' since the migration on the server
   * side already advanced the user to the correct step.
   */
  get currentStep(): 'eula' | 'permissions' | 'hotkey' | 'complete' {
    if (!this.daemon) return 'eula'
    const cs = this.daemon.current_step
    // Map the server's converged step names to the GUI union.
    // All four are valid directly. Legacy names are handled
    // by the daemon's migration; the GUI just trusts the
    // server's current_step.
    switch (cs) {
      case 'eula':
      case 'permissions':
      case 'hotkey':
      case 'complete':
        return cs
      default:
        // Legacy step name (welcome, power_source,
        // backend_detect, voice_test) — the daemon should
        // have already migrated. Default to eula so the
        // GUI shows the first step.
        return 'eula'
    }
  }

  /**
   * True when the wizard is complete. The GUI uses this to
   * dismiss the wizard and mount the main UI.
   */
  get isComplete(): boolean {
    if (!this.daemon) return false
    if (this.daemon.completed_at) return true
    return this.daemon.steps?.complete?.status === 'complete'
  }

  /**
   * Fetches the current wizard state from the daemon. Called
   * on mount and after every mutation.
   */
  async sync(): Promise<void> {
    this.busy = true
    this.error = null
    try {
      const state = await ipc.onboardingState()
      this.daemon = state
    } catch (e) {
      this.error = String(e)
    } finally {
      this.busy = false
    }
  }

  /**
   * Fetches the EULA document. The wizard shows the full
   * text in a scrollable area; the user must accept to
   * proceed. The accepted version is stored in the step
   * metadata so future EULA bumps force re-accept.
   */
  async loadEula(): Promise<void> {
    this.error = null
    try {
      const doc = await ipc.onboardingEula()
      this.eula = doc
      this.eulaVersion = doc.version
    } catch (e) {
      this.error = String(e)
    }
  }

  /**
   * Accepts the EULA. The version is recorded in step
   * metadata (so future EULA bumps force re-accept), then
   * the wizard advances to permissions. Re-syncs on
   * completion.
   */
  async acceptEula(version?: string): Promise<void> {
    this.busy = true
    this.error = null
    try {
      const v = version ?? this.eulaVersion ?? 'v1'
      // Record the accepted version in step metadata.
      await ipc.onboardingSetStep('eula', 'complete', v)
      // Advance to permissions.
      await ipc.onboardingAdvance()
      await this.sync()
    } catch (e) {
      this.error = String(e)
    } finally {
      this.busy = false
    }
  }

  /**
   * Records the EULA accept without advancing (for cases
   * where the wizard uses skipStep instead).
   */
  async back(): Promise<void> {
    this.busy = true
    this.error = null
    try {
      await ipc.onboardingBack()
      await this.sync()
    } catch (e) {
      this.error = String(e)
    } finally {
      this.busy = false
    }
  }

  /**
   * Marks the permissions step complete and advances to
   * hotkey. Called when the user has granted the required
   * OS permissions.
   */
  async completePermissions(): Promise<void> {
    this.busy = true
    this.error = null
    try {
      await ipc.onboardingSetStep('permissions', 'complete')
      await ipc.onboardingAdvance()
      await this.sync()
    } catch (e) {
      this.error = String(e)
    } finally {
      this.busy = false
    }
  }

  /**
   * Skips a step (e.g. the user opts out of granting
   * permissions). The daemon marks the step as skipped and
   * advances to the next one.
   */
  async skipStep(step: 'permissions' | 'hotkey'): Promise<void> {
    this.busy = true
    this.error = null
    try {
      await ipc.onboardingSkip(step)
      await this.sync()
    } catch (e) {
      this.error = String(e)
    } finally {
      this.busy = false
    }
  }

  /**
   * Sets the user's chosen hotkey. Persists only when
   * finish() is called; this just records the value in
   * the store so the GUI can show it in the hotkey step.
   */
  setHotkey(combo: string): void {
    this.hotkeyValue = combo
  }

  /**
   * Records the hotkey step as complete and advances to
   * complete. Called by the hotkey screen when the user
   * clicks Continue.
   */
  async saveHotkey(): Promise<void> {
    this.busy = true
    this.error = null
    try {
      await ipc.onboardingSetStep('hotkey', 'complete', this.hotkeyValue)
      await ipc.onboardingAdvance()
      await this.sync()
    } catch (e) {
      this.error = String(e)
    } finally {
      this.busy = false
    }
  }

  /**
   * Runs the power probe (Ollama + CLI detection). The
   * Ready screen shows the result; the user picks a power
   * source from the wizard.
   */
  async probePower(): Promise<void> {
    this.error = null
    try {
      this.power = await ipc.onboardingProbePower()
    } catch (e) {
      this.error = String(e)
    }
  }

  /**
   * Resets the wizard back to step 1. Called from Settings
   * when the user wants to re-run the wizard (e.g. to grant
   * a permission they previously skipped).
   */
  async reset(): Promise<void> {
    this.busy = true
    this.error = null
    try {
      await ipc.onboardingReset()
      this.daemon = null
      this.hotkeyValue = ''
      this.eula = null
      this.eulaVersion = ''
      this.power = null
      await this.sync()
    } catch (e) {
      this.error = String(e)
    } finally {
      this.busy = false
    }
  }

  /**
   * Completes the wizard. The daemon persists the hotkey
   * + telemetry choice, marks onboarding complete, and
   * returns the final state for the Ready screen.
   */
  async finish(
    params: Partial<Omit<OnboardingFinishParams, 'hotkey' | 'eula_version'>> & {
      eula_version?: string
      hotkey?: string
    }
  ): Promise<OnboardingFinishResult | null> {
    this.busy = true
    this.error = null
    try {
      const fullParams: OnboardingFinishParams = {
        hotkey: params.hotkey ?? this.hotkeyValue,
        eula_version: params.eula_version ?? this.eulaVersion ?? 'v1',
        permissions_skipped: params.permissions_skipped,
      }
      const result = await ipc.onboardingFinish(fullParams)
      await this.sync()
      return result
    } catch (e) {
      this.error = String(e)
      return null
    } finally {
      this.busy = false
    }
  }
}

// Singleton instance — only one wizard is in flight at a time.
export const onboarding = new OnboardingStore()
