// Auto-focus helper — three Meridian call sites (Sync PIN, Chat
// composer, Channels token input) had near-identical focus + select
// logic. Extract it so future "the user came here to do one thing"
// fields use the same shape.

export interface FocusOnOptions {
  /** Select the input's existing text after focusing. Defaults true
   *  because the canonical use case is "user came to overwrite what's
   *  already here." Set false for fields that should place the cursor
   *  in an empty value. */
  select?: boolean
  /** Don't scroll the page on focus. Defaults true because the
   *  focus change is usually a transition between sibling regions,
   *  not a "find the field" operation. */
  preventScroll?: boolean
}

/**
 * Focus + optionally select a ref-bound element when a condition
 * becomes true. Designed to be called from inside a Svelte `$effect`
 * for reactive triggers, or directly after an async transition for
 * imperative ones.
 *
 * The work is queued to a microtask so it runs AFTER Svelte's pending
 * DOM updates from the same tick (e.g. an `{#if}` flipping on the
 * same change that triggered the effect).
 *
 *   $effect(() => {
 *     focusOn(() => pinEl, () => !!sync.pendingPin && !pinExpired)
 *   })
 *
 *   async function openThread(id) {
 *     await conversation.open(id)
 *     focusOn(() => ta)
 *   }
 */
export function focusOn(
  ref: () => HTMLElement | null | undefined,
  when: () => boolean = () => true,
  options: FocusOnOptions = {}
): void {
  if (!when()) return
  const { select = true, preventScroll = true } = options
  queueMicrotask(() => {
    const el = ref()
    if (!el) return
    el.focus({ preventScroll })
    if (select && typeof (el as HTMLInputElement).select === 'function') {
      ;(el as HTMLInputElement).select()
    }
  })
}