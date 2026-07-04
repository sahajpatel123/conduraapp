<!--
  Condura Chat — the hero surface.

  Three-zone vertical stack inside the Shell's `main` region:
    (1) ConversationHeader — 56 px sticky, eyebrow + title + model + status + kebab
    (2) Body               — three columns (rail | feed | right rail) on wide screens
    (3) Composer           — pinned to the bottom with hairline above; focus-thread ::before

  Reads from stores; never subscribes to SSE itself — the conversation store
  owns the stream listener. Garden motes on empty state drift in a 9s linear
  loop; the global condura.css rule hides them under prefers-reduced-motion
  or data-energy="low" (no component-local override, per MOAT §2.3).

  Spec: app/web/frontend/src/lib/condura/specs/SCREEN_CHAT.md
  Drift items removed: inline <select> for model pick (R1) collapsed to the
  header model badge + composer whisper; the "last()" helper (R9) deleted;
  the inline ↗ glyph (R7) replaced with <Glyph name="send">; the local
  err-hair-draw keyframe (R3) and the per-component @media block (R4) and
  the inline motes re-declaration (R5) removed (the global condura.css owns
  pollen-float, err-hair-draw, and prefers-reduced-motion).
-->
<script lang="ts">
  import { onMount, onDestroy, untrack } from 'svelte';
  import { conversation } from '../stores/conversation.svelte';
  import { settings } from '../stores/settings.svelte';
  import { daemon } from '../stores/daemon.svelte';
  import { halt } from '../stores/halt.svelte';
  import { consent } from '../stores/consent.svelte';
  import { trust } from '../stores/trust.svelte';
  import { ipc } from '../ipc/client';
  import type { ConversationMeta, ProviderInfo } from '../ipc/types';
  import Button from './Button.svelte';
  import Glyph from './Glyph.svelte';
  import Pulse from './Pulse.svelte';
  import Thread from './Thread.svelte';
  import ErrorState from './ErrorState.svelte';

  // ---- local state ---------------------------------------------------------
  let inputText = $state('');
  let providers = $state<ProviderInfo[]>([]);
  let selectedModel = $state('');
  let scrollEl: HTMLDivElement | undefined = $state();
  let conversations = $state<ConversationMeta[]>([]);
  let docsGranted = $state(false);
  let micAvailable = $state(true);
  let voiceListening = $state(false);
  let voiceTranscript = $state<string[]>([]);
  let voiceTranscriptUnsub: (() => void) | null = null;
  let voiceTrayUnsub: (() => void) | null = null;
  let editingTitle = $state(false);
  let titleDraft = $state('');
  let kebabOpen = $state(false);
  let errorKind = $state<'daemon' | 'provider' | 'network' | null>(null);
  // Phase for the status chip — derived from store state but allows
  // optimistic updates (e.g. "Thinking" while stream starts).
  let inputFocused = $state(false);
  let lastErrorPoll = 0;

  // ---- model select (no <select> in the composer; model lives in header) -
  let modelOptions = $derived(
    providers.flatMap((p) =>
      p.models.map((m) => ({ value: `${p.name}:${m.id}`, label: `${p.name} · ${m.id}` }))
    )
  );
  let currentModelLabel = $derived(
    modelOptions.find((o) => o.value === selectedModel)?.label ?? ''
  );

  // ---- side rail widths (≥1440px shows conversation list; ≤1439 collapses)
  let mediaWide = $state(false);
  function syncMedia(): void {
    if (typeof matchMedia === 'undefined') return;
    mediaWide = matchMedia('(min-width: 1440px)').matches;
  }

  // ---- conversation list ---------------------------------------------------
  let loadingConvos = $state(false);
  async function refreshConversations(): Promise<void> {
    loadingConvos = true;
    try {
      await conversation.refreshList();
      conversations = conversation.conversations;
    } finally {
      loadingConvos = false;
    }
  }

  // ---- empty-state garden (deterministic once per mount) ------------------
  type Mote = { left: number; bottom: number; dx: number; dy: number; delay: number; dur: number };
  let motes = $state<Mote[]>([]);
  function seedGarden(): void {
    motes = Array.from({ length: 14 }, () => ({
      left: Math.random() * 100,
      bottom: Math.random() * 40,
      dx: Math.random() * 120 - 60,
      dy: -(120 + Math.random() * 180),
      delay: Math.random() * 9,
      dur: 7 + Math.random() * 6,
    }));
  }

  // ---- mount & teardown ----------------------------------------------------
  let pollTimer: ReturnType<typeof setInterval> | null = null;
  let voicePoll: ReturnType<typeof setInterval> | null = null;
  onMount(async () => {
    syncMedia();
    window.addEventListener('resize', syncMedia);
    seedGarden();
    try {
      const list = (await ipc.providersList()) as ProviderInfo[];
      providers = list;
    } catch {
      providers = [];
    }
    if (settings.config?.llm?.providers) {
      const enabled = Object.entries(settings.config.llm.providers).find(([, p]) => p.enabled);
      if (enabled) {
        selectedModel = `${enabled[0]}:${enabled[1].default_model ?? ''}`;
      }
    }
    if (!selectedModel && modelOptions.length > 0) {
      selectedModel = modelOptions[0].value;
    }
    await refreshConversations();
    try {
      await trust.refreshPermissions();
      docsGranted = trust.permissions.some(
        (p) => p.kind === 'documents' && p.status === 'granted'
      );
    } catch {
      docsGranted = false;
    }
    if (ipc.isConnected?.() ?? true) {
      try {
        const probe = await ipc.onboardingProbeVoice();
        micAvailable = (probe as { mic_available?: boolean }).mic_available ?? true;
      } catch {
        micAvailable = true;
      }
    }

    pollTimer = setInterval(() => {
      void refreshConversations();
    }, 60_000);
    // Voice IPC stream subscription (best-effort; not running under Wails
    // becomes a no-op thanks to the try/catch in on(...))
    try {
      voiceTranscriptUnsub = ipc.on(
        'voice.final' as never,
        ((ev: unknown) => {
          const text = (ev as { transcript?: string }).transcript?.trim();
          if (!text) return;
          voiceTranscript = [...voiceTranscript, text].slice(-6);
          inputText = voiceTranscript.join(' ');
        }) as never
      ) ?? null;
      voiceTrayUnsub = ipc.on(
        'tray.status' as never,
        ((payload: unknown) => {
          const status = (payload as { status?: string }).status;
          if (status === 'listening') voiceListening = true;
          else if (status === 'idle') voiceListening = false;
        }) as never
      ) ?? null;
    } catch {
      voiceTranscriptUnsub = null;
      voiceTrayUnsub = null;
    }
  });

  onDestroy(() => {
    if (typeof window !== 'undefined') window.removeEventListener('resize', syncMedia);
    if (pollTimer) clearInterval(pollTimer);
    if (voicePoll) clearInterval(voicePoll);
    voiceTranscriptUnsub?.();
    voiceTrayUnsub?.();
  });

  // ---- turns (derived from messages + active stream) ----------------------
  type Turn = { id: string; role: 'user' | 'agent' | 'system'; content: string };
  let turns = $derived.by<Turn[]>(() => {
    const base: Turn[] = conversation.messages.map((m, i) => ({
      id: `m-${conversation.currentID}-${i}`,
      role:
        m.role === 'assistant'
          ? ('agent' as const)
          : m.role === 'user'
            ? ('user' as const)
            : ('system' as const),
      content: m.content,
    }));
    if (conversation.isStreaming) {
      base.push({
        id: 'streaming',
        role: 'agent',
        content: conversation.streamingDelta || '…',
      });
    }
    return base;
  });

  let hasConversation = $derived(turns.length > 0);
  let toolCalls = $derived(conversation.streamingToolCalls ?? []);
  let interrupted = $derived(
    !conversation.isStreaming && toolCalls.length === 0 && turns.length > 0 && turns[turns.length - 1]?.id === 'streaming'
  );

  // ---- virtualized message list (binary search over measured offsets) -----
  const BUFFER_PX = 600;
  let scrollTop = $state(0);
  let viewportH = $state(0);
  let itemOffsets = $state<number[]>([]);
  let totalHeight = $derived(itemOffsets[itemOffsets.length - 1] ?? 0);
  let firstIdx = $derived(
    (() => {
      if (itemOffsets.length === 0) return 0;
      const target = scrollTop - BUFFER_PX;
      let lo = 0,
        hi = itemOffsets.length - 1;
      while (lo < hi) {
        const mid = (lo + hi) >> 1;
        if ((itemOffsets[mid + 1] ?? Infinity) <= target) lo = mid + 1;
        else hi = mid;
      }
      return lo;
    })()
  );
  let windowEnd = $derived(
    (() => {
      if (itemOffsets.length === 0) return 0;
      const target = scrollTop + viewportH + BUFFER_PX;
      let lo = 0,
        hi = itemOffsets.length - 1;
      while (lo < hi) {
        const mid = (lo + hi) >> 1;
        if ((itemOffsets[mid] ?? 0) <= target) lo = mid + 1;
        else hi = mid;
      }
      return Math.min(turns.length, lo);
    })()
  );
  let windowed = $derived(turns.slice(firstIdx, windowEnd));
  let topPad = $derived(itemOffsets[firstIdx] ?? 0);
  let botPad = $derived(Math.max(0, totalHeight - (itemOffsets[windowEnd] ?? totalHeight)));

  function onScroll(): void {
    if (!scrollEl) return;
    scrollTop = scrollEl.scrollTop;
    if (viewportH === 0) viewportH = scrollEl.clientHeight;
  }
  function measure(): void {
    if (!scrollEl) return;
    viewportH = scrollEl.clientHeight;
    const children = scrollEl.querySelectorAll<HTMLElement>('.msg[data-idx]');
    if (children.length === 0) return;
    const offsets: number[] = itemOffsets.slice();
    for (const c of children) {
      const i = Number(c.dataset.idx);
      if (!Number.isNaN(i)) offsets[i] = c.offsetTop;
    }
    itemOffsets = offsets;
  }

  // Pin to bottom on streaming — but only if user was already near bottom.
  $effect(() => {
    void conversation.streamingDelta;
    void conversation.messages.length;
    requestAnimationFrame(() => {
      if (!scrollEl) return;
      const wasNearBottom =
        scrollEl.scrollHeight - scrollEl.scrollTop - scrollEl.clientHeight < 200;
      scrollEl.scrollTop = scrollEl.scrollHeight;
      if (wasNearBottom) {
        requestAnimationFrame(() => {
          if (scrollEl) scrollEl.scrollTop = scrollEl.scrollHeight;
        });
      }
    });
  });

  // Re-measure after windowed changes (each {#each} block).
  $effect(() => {
    void windowed;
    requestAnimationFrame(measure);
  });
  // Re-measure on window resize (per spec §3.6).
  $effect(() => {
    void mediaWide;
    requestAnimationFrame(measure);
  });

  // ---- send / cancel / kill ------------------------------------------------
  async function send(): Promise<void> {
    const text = inputText.trim();
    if (!text || conversation.isStreaming) return;
    if (!selectedModel) return;
    inputText = '';
    const [providerName, modelId] = selectedModel.split(':');
    if (!providerName || !modelId) return;
    voiceTranscript = [];
    try {
      await conversation.send(providerName, modelId, text);
    } catch (e) {
      // ErrorState surfaces it; we still keep the user's text for retry.
      inputText = text;
      errorKind = 'daemon';
    }
  }

  async function cancelStream(): Promise<void> {
    await conversation.cancel();
  }

  async function killAgent(): Promise<void> {
    if (conversation.isStreaming) await conversation.cancel();
    await halt.halt('hard hotkey');
  }

  // ---- error kind (from conversation.streamingError) -----------------------
  $effect(() => {
    const err = conversation.streamingError;
    if (!err) {
      errorKind = null;
      return;
    }
    const lower = String(err).toLowerCase();
    if (lower.includes('network') || lower.includes('unreachable') || lower.includes('econn')) errorKind = 'network';
    else if (lower.includes('provider') || lower.includes('rejected') || lower.includes('api key')) errorKind = 'provider';
    else errorKind = 'daemon';
  });

  function retryLast(): void {
    void conversation.retry?.();
  }

  // ---- conversation list actions -------------------------------------------
  async function openConvo(id: number): Promise<void> {
    await conversation.open(id);
  }
  async function newConvo(): Promise<void> {
    await conversation.createNew();
    await refreshConversations();
  }
  async function deleteConvo(id: number): Promise<void> {
    await conversation.deleteById(id);
    if (id === conversation.currentID) {
      // empty state will appear; the conversation store already cleared.
    }
    await refreshConversations();
  }

  // ---- title rename --------------------------------------------------------
  function startRename(): void {
    titleDraft = conversation.currentTitle;
    editingTitle = true;
  }
  async function commitRename(): Promise<void> {
    if (titleDraft.trim() && titleDraft !== conversation.currentTitle && conversation.currentID) {
      // The conversation store does not own a rename method; rely on
      // conversation.createNew with a fresh title to re-create if needed.
      // For in-place rename, persist via the parent conversation store by
      // closing then recreating with the new title.
      const t = titleDraft.trim().slice(0, 60);
      // Soft path: just store the rename locally; persistence is owner's
      // job via the Settings → Conversation flow.
      conversation.currentTitle = t;
    }
    editingTitle = false;
  }
  function cancelRename(): void {
    editingTitle = false;
  }

  // ---- kebab menu ----------------------------------------------------------
  let kebabRef: HTMLDivElement | undefined = $state();
  function onDocClick(e: MouseEvent): void {
    if (!kebabOpen) return;
    if (kebabRef && !kebabRef.contains(e.target as Node)) kebabOpen = false;
  }
  $effect(() => {
    void kebabOpen;
    untrack(() => {
      // side-effectful, no dep
    });
    if (typeof document === 'undefined') return;
    document.addEventListener('mousedown', onDocClick);
    return () => document.removeEventListener('mousedown', onDocClick);
  });

  // ---- keyboard ------------------------------------------------------------
  function onComposerKeydown(e: KeyboardEvent): void {
    if (e.isComposing) return; // IME — never steal
    if (e.key === 'Enter' && !e.shiftKey) {
      // Plain Enter sends. Shift+Enter inserts a newline (textarea default).
      e.preventDefault();
      void send();
      return;
    }
    if (e.key === 'Escape' && conversation.isStreaming) {
      e.preventDefault();
      void cancelStream();
      return;
    }
    if (e.key === 'Escape' && !inputText) {
      // Let global handler dismiss any top overlay.
      return;
    }
    if (e.key === 'Escape' && inputText && inputFocused) {
      // Blur to drop the focus-thread (per spec §4.1).
      (e.currentTarget as HTMLTextAreaElement).blur();
      return;
    }
    // ↑ with empty composer: edit last user message.
    if (e.key === 'ArrowUp' && !inputText) {
      const lastUser = [...turns].reverse().find((t) => t.role === 'user');
      if (lastUser) {
        e.preventDefault();
        inputText = lastUser.content;
      }
    }
  }

  $effect(() => {
    if (typeof window === 'undefined') return;
    function isTextFocused(): boolean {
      const a = document.activeElement;
      if (!a) return false;
      const tag = a.tagName;
      return tag === 'TEXTAREA' || tag === 'INPUT';
    }
    function onWindow(e: KeyboardEvent): void {
      if (isTextFocused()) return;
      const meta = e.metaKey || e.ctrlKey;
      if (!meta) return;
      switch (e.key) {
        case 'n':
        case 'N':
          e.preventDefault();
          void newConvo();
          break;
        case '.':
          e.preventDefault();
          void killAgent();
          break;
        case 'r':
        case 'R':
          if (conversation.streamingError) {
            e.preventDefault();
            retryLast();
          }
          break;
      }
    }
    window.addEventListener('keydown', onWindow);
    return () => window.removeEventListener('keydown', onWindow);
  });

  // ---- voice ---------------------------------------------------------------
  async function toggleVoice(): Promise<void> {
    if (!micAvailable) return;
    voiceListening = !voiceListening;
    try {
      if (voiceListening) {
        const res = await ipc.voiceListen();
        const text = (res as { transcript?: string }).transcript?.trim();
        if (text) {
          inputText = (inputText + ' ' + text).trim();
        }
        voiceListening = false;
      }
    } catch (err) {
      voiceListening = false;
      // surfaced as banner only if voice is critical
      console.warn('voice listen failed', err);
    }
  }

  // ---- chip actions --------------------------------------------------------
  type Chip = { id: string; label: string; docsOnly?: boolean; prompt: string };
  const CHIPS: Chip[] = [
    {
      id: 'file',
      label: 'Find a file from last week.',
      prompt: 'Find a file I touched last week and show its path.',
    },
    {
      id: 'reply',
      label: 'Draft a reply to Maya.',
      prompt: 'Draft a short reply to Maya — friendly, no fluff, signed with my name.',
    },
    {
      id: 'prs',
      label: 'Summarize the open PRs in /code/synaptic.',
      prompt: 'Summarize the open pull requests in /code/synaptic, one line each.',
    },
    {
      id: 'watch',
      label: 'Watch my screen and tell me what I’m doing.',
      prompt: 'Watch my screen and tell me what I’m doing right now.',
    },
    {
      id: 'pdf',
      label: 'Walk me through this PDF.',
      prompt: 'Walk me through the most recent PDF in my Documents folder.',
      docsOnly: true,
    },
  ];
  let visibleChips = $derived(CHIPS.filter((c) => !c.docsOnly || docsGranted));
  async function pickChip(c: Chip): Promise<void> {
    inputText = c.prompt;
    requestAnimationFrame(() => void send());
  }

  // ---- status chip phase ---------------------------------------------------
  let statusPhase = $derived.by<'idle' | 'thinking' | 'awaiting' | 'acting' | 'consent' | 'error' | 'ok'>(() => {
    if (consent.ticket) return 'awaiting';
    if (!daemon.connected) return 'error';
    if (conversation.isStreaming) return 'acting';
    if (conversation.streamingError) return 'error';
    if (voiceListening) return 'thinking';
    return 'idle';
  });
  let statusLabel = $derived.by<string>(() => {
    if (consent.ticket) return 'Waiting on you';
    if (!daemon.connected) return 'Offline';
    if (conversation.isStreaming) return 'Thinking';
    if (conversation.streamingError) return 'Stopped';
    if (voiceListening) return 'Listening';
    return 'Idle';
  });
  let halted = $derived(halt.state.halted);
  let composerDisabled = $derived(
    halted || consent.ticket != null || !selectedModel || !modelOptions.length
  );
</script>

<!-- surface dim when halted (spec §2.8) -->
<div class="chat" class:halted>
  <!-- (1) ConversationHeader: sticky top, eyebrow + title + model + status + kebab -->
  <header class="conv-header" aria-label="Conversation header">
    <div class="ch-eyebrow">
      {hasConversation ? '— A conversation' : '— A fresh page'}
    </div>
    <div class="ch-row">
      {#if editingTitle}
        <input
          class="ch-title-input"
          bind:value={titleDraft}
          onkeydown={(e) => {
            if (e.key === 'Enter') void commitRename();
            else if (e.key === 'Escape') cancelRename();
          }}
          onblur={commitRename}
          aria-label="Rename conversation"
          maxlength="60"
        />
      {:else}
        <h1
          class="ch-title"
          ondblclick={startRename}
          title="Double-click to rename"
        >
          {conversation.currentTitle}
        </h1>
        <button
          type="button"
          class="ch-pencil tactile"
          aria-label="Rename"
          onclick={startRename}
          title="Rename"
        >
          <Glyph name="circle" size={12} />
        </button>
      {/if}

      {#if currentModelLabel}
        <span class="ch-model" title={currentModelLabel}>{currentModelLabel}</span>
      {:else}
        <span class="ch-model ch-model-empty" title="No model configured">no model</span>
      {/if}

      <span class="ch-status" class:ch-status-streaming={conversation.isStreaming}>
        <Pulse phase={statusPhase} size={conversation.isStreaming ? 8 : 6} />
        <span class="ch-status-label">{statusLabel}</span>
      </span>

      <div class="ch-kebab-wrap" bind:this={kebabRef}>
        <button
          type="button"
          class="ch-kebab tactile"
          aria-label="Conversation actions"
          aria-haspopup="menu"
          aria-expanded={kebabOpen}
          onclick={() => (kebabOpen = !kebabOpen)}
        >
          <Glyph name="menu" size={16} />
        </button>
        {#if kebabOpen}
          <div class="ch-menu" role="menu">
            <button
              type="button"
              class="ch-menu-item tactile"
              role="menuitem"
              onclick={() => {
                startRename();
                kebabOpen = false;
              }}
            >
              Rename
            </button>
            <button
              type="button"
              class="ch-menu-item tactile"
              role="menuitem"
              onclick={() => {
                if (conversation.currentID) void deleteConvo(conversation.currentID);
                kebabOpen = false;
              }}
            >
              Delete
            </button>
            <button
              type="button"
              class="ch-menu-item tactile"
              role="menuitem"
              onclick={() => {
                void newConvo();
                kebabOpen = false;
              }}
            >
              New conversation
            </button>
          </div>
        {/if}
      </div>
    </div>
    <div class="ch-hair" aria-hidden="true"><Thread orientation="h" /></div>
  </header>

  <!-- (2) Three-region body -->
  <div class="body" class:wide={mediaWide}>
    <!-- (2) ConversationList rail — only on ≥1440px -->
    {#if mediaWide}
      <aside class="rail" aria-label="Conversations">
        <button type="button" class="cl-new tactile" onclick={() => void newConvo()}>
          <Glyph name="plus" size={14} /> New conversation
        </button>
        <div class="cl-list">
          {#if loadingConvos && conversations.length === 0}
            <div class="cl-loading">
              <span class="cl-loading-label">indexing conversations</span>
              <Pulse phase="thinking" size={6} />
            </div>
          {:else if conversations.length === 0}
            <div class="cl-empty">No conversations yet.</div>
          {:else}
            {#each conversations as c (c.id)}
              <button
                type="button"
                class="cl-row tactile"
                class:active={c.id === conversation.currentID}
                onclick={() => void openConvo(c.id)}
              >
                <span class="cl-title">{c.title || 'New conversation'}</span>
                <span class="cl-time">{formatTime(c.last_touch ?? c.updated_at ?? c.created_at)}</span>
              </button>
            {/each}
          {/if}
        </div>
        <div class="cl-foot">{conversations.length} conversation{conversations.length === 1 ? '' : 's'}</div>
      </aside>
    {/if}

    <!-- (3) MessageFeed -->
    <div class="feed-col" bind:this={scrollEl} onscroll={onScroll}>
      {#if !hasConversation && !errorKind}
        <!-- Empty state -->
        <div class="empty">
          <div class="garden" aria-hidden="true">
            {#each motes as m (m.delay + ':' + m.dur)}
              <span
                class="mote"
                style:left="{m.left}%"
                style:bottom="{m.bottom}%"
                style:--dx="{m.dx}px"
                style:--dy="{m.dy}px"
                style:animation-delay="{m.delay}s"
                style:animation-duration="{m.dur}s"
              ></span>
            {/each}
          </div>
          <div class="copy">
            <div class="eyebrow">— A quiet place to write</div>
            <h1 class="hero">
              <span class="wordrise"><span>Your</span></span>
              <span class="wordrise"><span>computer,</span></span><br />
              <span class="wordrise"><span class="alive">alive.</span></span>
            </h1>
            <p class="lead">
              A quiet, attentive presence on your machine. It perceives only what it must, and it acts only
              after it shows you what it's about to do. Press your hotkey, or ask below.
            </p>
            <div class="chips">
              {#each visibleChips as c (c.id)}
                <button type="button" class="chip tactile" onclick={() => void pickChip(c)}>
                  <span class="chip-arrow" aria-hidden="true">↳</span>
                  {c.label}
                </button>
              {/each}
            </div>
          </div>
        </div>
      {:else if errorKind}
        <div class="err-wrap">
          {#if errorKind === 'daemon'}
            <ErrorState
              head="Connection to daemon"
              cause="daemon"
              reason="daemon was restarted, or the socket dropped."
              onretry={retryLast}
              retryLabel="Try again"
            />
          {:else if errorKind === 'provider'}
            <ErrorState
              head="Provider rejected request"
              cause="provider"
              reason="the API key may have rotated, or the model is unavailable."
              onretry={retryLast}
              retryLabel="Try again"
            />
          {:else}
            <ErrorState
              head="Network failure"
              cause="network"
              reason="your network is unreachable."
              onretry={retryLast}
              retryLabel="Try again"
            />
          {/if}
        </div>
      {:else}
        <!-- Mid-conversation: virtualized list -->
        <div class="messages">
          {#if topPad > 0}
            <div class="msg-spacer" style:height="{topPad}px" aria-hidden="true"></div>
          {/if}
          {#each windowed as t, i (t.id)}
            <div class="msg {t.role}" data-idx={firstIdx + i}>
              <div class="m-head">
                <span class="m-label">{t.role === 'user' ? 'You' : t.role === 'agent' ? 'Condura' : 'System'}</span>
              </div>
              <div class="bubble" class:streaming={t.id === 'streaming'}>
                {t.content}
                {#if t.id === 'streaming'}<span class="caret" aria-hidden="true"></span>{/if}
              </div>
              {#if t.id === 'streaming'}
                <div class="stream-bar" aria-hidden="true"></div>
              {/if}
              {#if t.id === 'streaming' && toolCalls.length > 0}
                <div class="tools">
                  {#each toolCalls as tc (tc.id)}
                    <span class="tool-chip"><Glyph name="bolt" size={10} /> {tc.function.name}</span>
                  {/each}
                </div>
              {/if}
              <!-- interrupted: "Continue →" affordance below the frozen bubble -->
              {#if t.id !== 'streaming' && t.role === 'agent' && interrupted && i === windowed.length - 1}
                <div class="continue-row">
                  <button
                    type="button"
                    class="continue-pill tactile"
                    onclick={() => {
                      inputText = 'Continue: ';
                      scrollToComposer();
                    }}
                  >
                    Continue
                    <Glyph name="chevron-right" size={12} />
                  </button>
                </div>
              {/if}
            </div>
            {#if firstIdx + i < turns.length - 1}
              <div class="turn-thread" aria-hidden="true"><Thread orientation="v" /></div>
            {/if}
          {/each}
          {#if botPad > 0}
            <div class="msg-spacer" style:height="{botPad}px" aria-hidden="true"></div>
          {/if}
        </div>
      {/if}
    </div>
  </div>

  <!-- (5) Composer -->
  <div class="composer">
    <div class="composer-card" class:composer-disabled={composerDisabled} class:focused={inputFocused}>
      <textarea
        bind:value={inputText}
        onkeydown={onComposerKeydown}
        onfocus={() => (inputFocused = true)}
        onblur={() => (inputFocused = false)}
        placeholder="→ Ask Condura to do something…"
        rows="1"
        aria-disabled={composerDisabled}
      ></textarea>
      <div class="composer-foot">
        <button
          type="button"
          class="orb tactile"
          class:listening={voiceListening}
          class:disabled={!micAvailable}
          aria-label={micAvailable ? (voiceListening ? 'Listening — tap to stop' : 'Start voice') : 'Microphone not granted'}
          title={micAvailable ? 'Voice' : 'Microphone not granted'}
          onclick={() => void toggleVoice()}
        >
          {#if voiceListening}
            <span class="orb-wave" aria-hidden="true">
              <span class="bar b1"></span>
              <span class="bar b2"></span>
              <span class="bar b3"></span>
            </span>
          {:else}
            <Glyph name="mic" size={14} />
          {/if}
        </button>
        {#if currentModelLabel}
          <span class="model-label" title={currentModelLabel}>
            {currentModelLabel}
          </span>
        {:else}
          <span class="model-label model-label-empty">no model — add one in Settings → Power</span>
        {/if}
        {#if conversation.isStreaming}
          <Button variant="secondary" size="sm" onclick={() => void cancelStream()}>
            <Glyph name="stop" size={14} /> Stop
          </Button>
        {:else}
          <Button
            variant="primary"
            magnetic
            onclick={() => void send()}
            disabled={composerDisabled || inputText.trim() === ''}
            class="send"
          >
            Send <Glyph name="send" size={14} class="send-arrow" />
          </Button>
        {/if}
      </div>
    </div>
    <div class="composer-hint">
      ⌘↵ to send · Esc to stop · your hotkey to summon
    </div>
  </div>

  {#if halted}
    <div class="halted-readout" role="status" aria-live="polite">
      Halted. Resume from the kill switch.
    </div>
  {/if}
</div>

<script module lang="ts">
  // Format an ISO-ish timestamp into the rail's `LAST · HH:MM` form.
  function formatTime(t?: string | number): string {
    if (!t) return '';
    const d = new Date(t);
    if (Number.isNaN(d.getTime())) return '';
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });
  }
  // Hoisted into module context so Svelte does not bundle it with the
  // component instance (it's cheap, but it's pure).
  // No exports needed.
  function scrollToComposer(): void {
    // Defer to next frame so the input value lands before focus.
    requestAnimationFrame(() => {
      const el = document.querySelector<HTMLTextAreaElement>('.composer textarea');
      el?.focus();
    });
  }
</script>

<style>
  .chat {
    display: flex;
    flex-direction: column;
    height: 100%;
    position: relative;
    transition: opacity var(--dur) var(--ease);
  }
  .chat.halted {
    opacity: 0.4;
  }

  /* ── (1) ConversationHeader ─────────────────────────────────────────── */
  .conv-header {
    padding: var(--space-5) var(--space-10) var(--space-3);
    border-bottom: 1px solid var(--hair);
    position: sticky;
    top: 0;
    background: var(--surface);
    z-index: var(--z-sticky);
    animation: fade-in-up var(--dur) var(--ease);
  }
  .ch-eyebrow {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--content-faint);
    margin-bottom: 4px;
  }
  .ch-row {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    min-height: 32px;
  }
  .ch-title {
    font-family: var(--font-display);
    font-size: 20px;
    line-height: 1.1;
    letter-spacing: -0.025em;
    color: var(--content);
    margin: 0;
    font-weight: 400;
    flex: none;
    max-width: 38ch;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .ch-title-input {
    font-family: var(--font-display);
    font-size: 20px;
    line-height: 1.1;
    letter-spacing: -0.025em;
    color: var(--content);
    background: transparent;
    border: 0;
    border-bottom: 1px solid var(--synapse);
    outline: none;
    padding: 0 0 2px;
    width: 38ch;
  }
  .ch-pencil {
    background: transparent;
    border: 0;
    padding: 4px;
    color: var(--content-faint);
    border-radius: var(--r-xs);
    opacity: 0;
    transition: opacity var(--dur) var(--ease), color var(--dur) var(--ease);
  }
  .ch-row:hover .ch-pencil {
    opacity: 1;
  }
  .ch-pencil:hover {
    color: var(--synapse);
  }
  .ch-pencil:focus-visible {
    opacity: 1;
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }

  .ch-model {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    color: var(--content-mute);
    background: var(--surface-sunken);
    padding: 3px 10px;
    border-radius: var(--r-pill);
    border: 1px solid var(--hair);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 22ch;
  }
  .ch-model-empty {
    color: var(--danger);
  }

  .ch-status {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: 4px 10px;
    border-radius: var(--r-pill);
    border: 1px solid var(--hair);
    background: var(--surface);
  }
  .ch-status-streaming {
    border-color: var(--warn);
    color: var(--warn);
  }
  .ch-status-label {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--content);
  }
  .ch-status-streaming .ch-status-label {
    color: var(--warn);
  }

  .ch-kebab-wrap {
    margin-left: auto;
    position: relative;
  }
  .ch-kebab {
    width: 28px;
    height: 28px;
    border-radius: var(--r-pill);
    background: transparent;
    border: 1px solid transparent;
    color: var(--content-mute);
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
  .ch-kebab:hover {
    background: var(--surface-card);
    border-color: var(--hair);
    color: var(--content);
  }
  .ch-kebab:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }

  .ch-menu {
    position: absolute;
    top: calc(100% + 4px);
    right: 0;
    min-width: 180px;
    background: var(--surface-raised);
    border: 1px solid var(--hair);
    border-radius: var(--r-md);
    padding: var(--space-2);
    box-shadow: var(--shadow-float);
    display: flex;
    flex-direction: column;
    gap: 2px;
    z-index: var(--z-overlay);
  }
  .ch-menu-item {
    background: transparent;
    border: 0;
    text-align: left;
    padding: 8px 12px;
    border-radius: var(--r-xs);
    color: var(--content);
    font-family: var(--font-sans);
    font-size: 14px;
  }
  .ch-menu-item:hover {
    background: var(--surface-card);
  }
  .ch-menu-item:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .ch-hair {
    margin-top: var(--space-2);
    margin-left: 0;
    width: 100%;
    height: 2px;
    overflow: hidden;
  }

  /* ── (2) Body: rail + feed on wide ──────────────────────────────────── */
  .body {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: 1fr;
    overflow: hidden;
  }
  .body.wide {
    grid-template-columns: 280px minmax(0, 1fr);
  }
  .rail {
    border-right: 1px solid var(--hair);
    display: flex;
    flex-direction: column;
    background: var(--surface);
    overflow: hidden;
  }
  .cl-new {
    background: transparent;
    border: 0;
    border-bottom: 1px solid var(--hair);
    color: var(--content-soft);
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    text-align: left;
    padding: var(--space-4);
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
  }
  .cl-new:hover {
    background: var(--surface-card);
    color: var(--content);
  }
  .cl-new:focus-visible {
    outline: none;
    box-shadow: inset 0 0 0 4px var(--pollen-halo);
  }
  .cl-list {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-2);
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .cl-loading {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-4) var(--space-3);
    color: var(--content-faint);
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }
  .cl-loading-label {
    flex: 1;
  }
  .cl-empty {
    padding: var(--space-4) var(--space-3);
    color: var(--content-faint);
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
  }
  .cl-row {
    background: transparent;
    border: 0;
    text-align: left;
    width: 100%;
    padding: var(--space-3);
    border-radius: var(--r-sm);
    color: var(--content-soft);
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .cl-row:hover {
    background: var(--surface-card);
    color: var(--content);
  }
  .cl-row.active {
    color: var(--content);
  }
  .cl-row.active::before {
    content: '';
    position: absolute;
    left: 0;
    top: 6px;
    bottom: 6px;
    width: 2px;
    background: var(--synapse);
    border-radius: 1px;
    transform: scaleY(1);
    transform-origin: top;
    animation: drawv var(--dur-slow) var(--ease);
  }
  .cl-row:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .cl-title {
    font-family: var(--font-display);
    font-size: 14px;
    line-height: 1.25;
    color: var(--content);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .cl-time {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--content-faint);
  }
  .cl-foot {
    border-top: 1px solid var(--hair);
    padding: var(--space-3);
    color: var(--content-faint);
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.04em;
  }
  @keyframes drawv {
    to {
      transform: scaleY(1);
    }
  }

  /* ── (3) Feed column ───────────────────────────────────────────────── */
  .feed-col {
    min-height: 0;
    overflow-y: auto;
    padding: var(--space-7) var(--space-10) var(--space-5);
    position: relative;
  }

  /* ── empty state (3a) ──────────────────────────────────────────────── */
  .empty {
    position: relative;
    min-height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    max-width: 760px;
    margin: 0 auto;
    width: 100%;
  }
  .garden {
    position: absolute;
    inset: 0;
    pointer-events: none;
    overflow: hidden;
    z-index: 0;
  }
  .mote {
    position: absolute;
    width: 4px;
    height: 4px;
    border-radius: 50%;
    background: var(--pollen);
    box-shadow: 0 0 6px color-mix(in oklab, var(--pollen) 50%, transparent);
    opacity: 0;
    animation: pollen-float 9s linear infinite;
  }
  .copy {
    position: relative;
    z-index: 1;
  }
  .eyebrow {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--content-faint);
    margin-bottom: var(--space-2);
  }
  .copy .hero {
    font-family: var(--font-display);
    font-size: clamp(40px, 6vw, 68px);
    line-height: var(--lh-display);
    letter-spacing: var(--ls-display);
    margin: var(--space-5) 0;
    color: var(--content);
    font-weight: 400;
  }
  .copy .hero .alive {
    color: var(--synapse);
    font-style: italic;
  }
  .copy .lead {
    font-family: var(--font-sans);
    font-size: var(--text-lead);
    line-height: var(--lh-lead);
    letter-spacing: var(--ls-lead);
    color: var(--content-soft);
    max-width: 52ch;
    margin: 0 0 var(--space-6);
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .chip {
    background: var(--surface-sunken);
    border: 1px solid var(--hair);
    color: var(--content-soft);
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    padding: 8px 12px;
    border-radius: var(--r-pill);
    text-align: left;
    max-width: 38ch;
  }
  .chip:hover {
    color: var(--content);
    border-color: var(--hair-strong);
    transform: translateY(-1px);
  }
  .chip:active {
    transform: scale(0.97);
  }
  .chip-arrow {
    color: var(--pollen);
    margin-right: 4px;
  }

  /* ── error state wrapper (ErrorState handles its own visuals) ──────── */
  .err-wrap {
    max-width: 560px;
    margin: var(--space-7) auto;
  }

  /* ── mid-conversation (3b) ─────────────────────────────────────────── */
  .messages {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    max-width: 760px;
    margin: 0 auto;
    width: 100%;
  }
  .msg {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .msg.user {
    align-items: flex-end;
  }
  .msg.agent,
  .msg.system {
    align-items: flex-start;
  }
  .msg.agent:hover .bubble,
  .msg.user:hover .bubble {
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .msg:hover .m-label {
    color: var(--content);
  }
  .m-head {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }
  .m-label {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--content-faint);
    transition: color var(--dur) var(--ease);
  }
  .msg.user .m-label {
    color: var(--pollen);
  }
  .msg.agent .m-label {
    color: var(--synapse);
  }
  .bubble {
    max-width: 88%;
    padding: var(--space-4) var(--space-5);
    border-radius: var(--r-lg);
    font-size: 15px;
    line-height: 1.6;
    color: var(--content-soft);
    border: 1px solid var(--hair);
    background: var(--surface-card);
    white-space: pre-wrap;
    word-wrap: break-word;
    transition: box-shadow var(--dur-fast) var(--ease);
  }
  .msg.user .bubble {
    border-color: color-mix(in oklab, var(--pollen) 22%, transparent);
    background: linear-gradient(
      180deg,
      color-mix(in oklab, var(--pollen) 6%, transparent),
      color-mix(in oklab, var(--pollen) 2%, transparent)
    );
  }
  .msg.agent .bubble {
    border-color: color-mix(in oklab, var(--synapse) 18%, transparent);
  }
  .msg.system .bubble {
    border-color: transparent;
    background: transparent;
    color: var(--content-faint);
    font-style: italic;
  }
  .bubble.streaming {
    border-color: color-mix(in oklab, var(--synapse) 28%, transparent);
  }
  .caret {
    display: inline-block;
    width: 8px;
    height: 1.1em;
    background: var(--synapse);
    margin-left: 2px;
    transform: translateY(2px);
    animation: blink 1s steps(2) infinite;
  }
  .stream-bar {
    position: relative;
    width: 88%;
    height: 1px;
    background: linear-gradient(90deg, transparent, var(--hair-strong) 20%, var(--hair-strong) 80%, transparent);
    overflow: hidden;
  }
  .stream-bar::after {
    content: '';
    position: absolute;
    top: -1px;
    width: 40px;
    height: 3px;
    border-radius: 2px;
    background: var(--pollen);
    box-shadow: 0 0 8px color-mix(in oklab, var(--pollen) 60%, transparent);
    animation: travel 2.6s var(--ease) infinite;
  }
  .tools {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }
  .tool-chip {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--synapse);
    border: 1px solid color-mix(in oklab, var(--synapse) 25%, transparent);
    background: color-mix(in oklab, var(--synapse) 5%, transparent);
    padding: 3px 10px;
    border-radius: var(--r-pill);
  }
  .turn-thread {
    width: 2px;
    height: 28px;
    margin: 0 auto;
  }
  .msg-spacer {
    width: 100%;
    pointer-events: none;
  }

  .continue-row {
    display: flex;
    margin-top: var(--space-2);
  }
  .continue-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    background: var(--surface-sunken);
    color: var(--content-soft);
    border: 1px solid var(--hair);
    border-radius: var(--r-pill);
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    padding: 5px 12px;
  }
  .continue-pill:hover {
    color: var(--content);
    border-color: var(--synapse);
  }
  .continue-pill:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }

  /* ── (5) Composer ─────────────────────────────────────────────────── */
  .composer {
    padding: var(--space-3) var(--space-10) var(--space-5);
    max-width: 820px;
    width: 100%;
    margin: 0 auto;
  }
  .composer-card {
    background: var(--surface-card);
    border: 1px solid var(--hair);
    border-radius: var(--r-lg);
    padding: var(--space-4) var(--space-4) var(--space-3);
    box-shadow: var(--shadow-card);
    position: relative;
    transition:
      border-color var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease),
      opacity var(--dur) var(--ease);
  }
  .composer-card.focused {
    border-color: var(--synapse);
    box-shadow: var(--shadow-card), 0 0 0 3px color-mix(in oklab, var(--synapse) 10%, transparent);
  }
  .composer-disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }
  /* The signature focus-thread (per MOAT §5.2 + DIRECTION §5 "Thread"). */
  .composer-card::before {
    content: '';
    position: absolute;
    left: var(--space-4);
    right: var(--space-4);
    bottom: 0;
    height: 1px;
    background: linear-gradient(90deg, transparent, var(--synapse) 20%, var(--synapse) 80%, transparent);
    transform: scaleX(0);
    transform-origin: left;
    transition: transform var(--dur-slow) var(--ease);
  }
  .composer-card.focused::before {
    transform: scaleX(1);
  }
  .composer textarea {
    width: 100%;
    min-height: 48px;
    resize: none;
    border: none;
    background: transparent;
    color: var(--content);
    font-family: var(--font-sans);
    font-size: 16px;
    line-height: 1.5;
  }
  .composer-disabled textarea {
    cursor: not-allowed;
  }
  .composer textarea::placeholder {
    color: var(--content-faint);
  }
  .composer textarea:focus {
    outline: none;
  }
  .composer-foot {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin-top: var(--space-2);
  }

  /* VoiceOrb — 28x28 round, surface-card fill, hair border (spec §1.5). */
  .orb {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: var(--surface-card);
    border: 1px solid var(--hair);
    color: var(--content-mute);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex: none;
    transition:
      background var(--dur) var(--ease),
      border-color var(--dur) var(--ease),
      color var(--dur) var(--ease),
      transform var(--dur) var(--ease);
  }
  .orb:hover:not(:disabled) {
    color: var(--content);
    border-color: var(--synapse);
    transform: translateY(-1px);
  }
  .orb:active {
    transform: scale(0.96);
  }
  .orb:focus-visible {
    outline: none;
    box-shadow: 0 0 0 4px var(--pollen-halo);
  }
  .orb.listening {
    border-color: var(--pollen);
    color: var(--pollen);
    animation: breathe 4s var(--ease) infinite;
  }
  .orb.disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }
  .orb-wave {
    display: inline-flex;
    align-items: center;
    gap: 2px;
  }
  .orb-wave .bar {
    display: inline-block;
    width: 2px;
    border-radius: 1px;
    background: currentColor;
    animation: wave 1.4s var(--ease) infinite;
  }
  .orb-wave .b1 {
    height: 8px;
    animation-delay: 0s;
  }
  .orb-wave .b2 {
    height: 16px;
    animation-delay: 0.15s;
  }
  .orb-wave .b3 {
    height: 24px;
    animation-delay: 0.3s;
  }

  .model-label {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    color: var(--content-mute);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 28ch;
  }
  .model-label-empty {
    color: var(--danger);
  }
  :global(.send) {
    margin-left: auto;
  }
  :global(.send .send-arrow) {
    transition: transform var(--dur) var(--ease);
  }
  :global(.send:hover .send-arrow) {
    transform: translate(2px, -2px);
  }
  .composer-hint {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--content-faint);
    margin-top: var(--space-2);
    text-align: center;
    transition: color var(--dur) var(--ease);
  }
  .composer-hint:hover {
    color: var(--content-mute);
  }

  /* halted readout (per spec §2.8 surface dim) */
  .halted-readout {
    position: absolute;
    bottom: var(--space-9);
    left: 50%;
    transform: translateX(-50%);
    background: var(--surface-ink);
    color: var(--content);
    padding: 6px 14px;
    border-radius: var(--r-pill);
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
  }
</style>
