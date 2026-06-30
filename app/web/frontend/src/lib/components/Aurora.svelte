<script lang="ts">
  // The Aurora Field — the one living light system of IRIS.
  // Three large, heavily-blurred radial blobs drift on slow,
  // independent loops behind everything. Compositor-only motion
  // (transform/opacity), capped at three layers, frozen under
  // prefers-reduced-motion. A faint grain overlay kills banding.
  //
  // `dim` slows + dims the field (used while the Setup Console is
  // open, so the room quiets to focus on the console).
  interface Props {
    dim?: boolean
  }
  let { dim = false }: Props = $props()
</script>

<div class="aurora" class:dim aria-hidden="true">
  <span class="blob blob-a"></span>
  <span class="blob blob-b"></span>
  <span class="blob blob-c"></span>
  <span class="grain"></span>
</div>

<style>
  .aurora {
    position: fixed;
    inset: 0;
    z-index: var(--z-ambient);
    overflow: hidden;
    background: var(--void);
    pointer-events: none;
    transition: opacity var(--transition-slow);
  }

  .blob {
    position: absolute;
    width: 72vmax;
    height: 72vmax;
    border-radius: 50%;
    filter: blur(120px);
    will-change: transform;
  }
  .blob-a {
    top: -14vmax;
    left: -8vmax;
    opacity: 0.5;
    background: radial-gradient(circle, var(--aurora-1), transparent 60%);
    animation: aurora-drift-a 38s ease-in-out infinite;
  }
  .blob-b {
    bottom: -18vmax;
    left: 22vmax;
    opacity: 0.46;
    background: radial-gradient(circle, var(--aurora-2), transparent 62%);
    animation: aurora-drift-b 46s ease-in-out infinite;
  }
  .blob-c {
    bottom: -14vmax;
    right: -10vmax;
    opacity: 0.32;
    background: radial-gradient(circle, var(--aurora-3), transparent 60%);
    animation: aurora-drift-c 52s ease-in-out infinite;
  }

  /* Thinking state speeds the drift; acting pushes coral forward.
     Driven by data-agent-state on <html>. */
  :global(html[data-agent-state='thinking']) .blob-a { animation-duration: 17s; }
  :global(html[data-agent-state='thinking']) .blob-b { animation-duration: 21s; }
  :global(html[data-agent-state='thinking']) .blob-c { animation-duration: 24s; opacity: 0.42; }
  :global(html[data-agent-state='acting']) .blob-c { opacity: 0.5; }

  .grain {
    position: absolute;
    inset: 0;
    opacity: 0.035;
    mix-blend-mode: overlay;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='160' height='160'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='2' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E");
  }

  /* Quiet the room while a modal/console is focused. */
  .aurora.dim .blob-a { animation-duration: 76s; opacity: 0.3; }
  .aurora.dim .blob-b { animation-duration: 92s; opacity: 0.28; }
  .aurora.dim .blob-c { animation-duration: 104s; opacity: 0.2; }

  @media (prefers-reduced-motion: reduce) {
    .blob { animation: none !important; }
  }
</style>
