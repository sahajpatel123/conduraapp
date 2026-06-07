<script lang="ts">
  import { Ping, DaemonStatus } from '../wailsjs/go/main/App.js'

  let resultText: string = "Welcome to Synaptic. Type a name and click Ping."
  let daemonText: string = "Daemon: starting..."
  let name: string = ""

  function ping(): void {
    Ping(name).then(result => resultText = result)
  }

  function checkDaemon(): void {
    DaemonStatus().then((s) => {
      if (s.ready) {
        daemonText = `Daemon: ready @ ${s.addr}`
      } else {
        daemonText = "Daemon: not ready yet"
      }
    })
  }

  // Poll daemon status every 500ms while the page is open.
  setInterval(checkDaemon, 500)
  checkDaemon()
</script>

<main>
  <h1>Synaptic</h1>
  <p class="daemon">{daemonText}</p>

  <p>{resultText}</p>
  <div class="input-row">
    <input autocomplete="off" bind:value={name} class="input" id="name" type="text" placeholder="Your name" />
    <button class="btn" on:click={ping}>Ping</button>
  </div>
</main>

<style>
  main {
    margin: 0;
    padding: 48px 24px;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", system-ui, sans-serif;
    color: #e6e6e6;
    background: #121216;
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }
  h1 {
    font-size: 36px;
    font-weight: 600;
    margin: 0;
  }
  .daemon {
    font-size: 14px;
    color: #9aa0a6;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    margin: 0 0 24px 0;
  }
  .input-row {
    display: flex;
    gap: 8px;
  }
  .input {
    background: #1d1d22;
    border: 1px solid #2a2a31;
    color: #e6e6e6;
    padding: 10px 14px;
    border-radius: 8px;
    font-size: 14px;
    width: 240px;
  }
  .input:focus {
    outline: none;
    border-color: #5b8def;
  }
  .btn {
    background: #5b8def;
    color: white;
    border: none;
    padding: 10px 18px;
    border-radius: 8px;
    font-size: 14px;
    cursor: pointer;
  }
  .btn:hover {
    background: #4a7ade;
  }
</style>
