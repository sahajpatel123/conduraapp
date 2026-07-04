// Tiny typed event emitter. 60 lines, no deps, no class boilerplate.
//
// We don't need the full Node `events` API — just on/off/emit and
// auto-unsubscribe via the returned function. This is the entire
// event system in the GUI.

type AnyHandler = (...args: any[]) => void
type HandlerMap = Map<string, Set<AnyHandler>>

export class EventEmitter<T extends Record<string, any[]>> {
  private handlers: HandlerMap = new Map()

  on<E extends keyof T>(event: E, handler: (...args: T[E]) => void): () => void {
    let set = this.handlers.get(event as string)
    if (!set) {
      set = new Set()
      this.handlers.set(event as string, set)
    }
    set.add(handler as AnyHandler)
    return () => this.off(event, handler)
  }

  off<E extends keyof T>(event: E, handler: (...args: T[E]) => void): void {
    const set = this.handlers.get(event as string)
    if (set) {
      set.delete(handler as AnyHandler)
    }
  }

  emit<E extends keyof T>(event: E, ...args: T[E]): void {
    const set = this.handlers.get(event as string)
    if (!set) {
      return
    }
    for (const handler of set) {
      try {
        handler(...args)
      } catch (err) {
        // We intentionally swallow handler errors so one bad
        // listener doesn't break the chain. Log to console.error
        // for the developer.
        // eslint-disable-next-line no-console
        console.error(`[EventEmitter] handler for ${String(event)} threw:`, err)
      }
    }
  }

  removeAllListeners(event?: keyof T): void {
    if (event) {
      this.handlers.delete(event as string)
    } else {
      this.handlers.clear()
    }
  }
}
