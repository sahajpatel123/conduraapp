<!--
  V2HubDemo — Condura v2 Hub preview.

  Skills as real book spines. Hover tilts them 4°. Loaded skills
  have an accent ribbon at the top of the spine. Search filters
  the back catalog instantly.
-->
<script lang="ts">
  import '$lib/v2/tokens.css'
  import '$lib/v2/motion.css'
  import '$lib/v2/reset.css'
  import { Sidebar, StatusBar, Hub, type SidebarItem, type Skill } from '$lib/v2'

  const items: SidebarItem[] = [
    { id: 'chat',       monogram: 'Ch', label: 'Chat' },
    { id: 'settings',   monogram: 'St', label: 'Settings' },
    { id: 'audit',      monogram: 'Au', label: 'Audit' },
    { id: 'channels',   monogram: 'Co', label: 'Channels' },
    { id: 'delegation', monogram: 'De', label: 'Delegation' },
    { id: 'hub',        monogram: 'Hu', label: 'Hub' },
    { id: 'replay',     monogram: 'Re', label: 'Replay' },
    { id: 'sync',       monogram: 'Sy', label: 'Sync' },
    { id: 'skills',     monogram: 'Sk', label: 'Skills' },
    { id: 'about',      monogram: 'Ab', label: 'About' },
  ]

  let active = $state('hub')
  let collapsed = $state(false)
  let query = $state('')

  let installed = $state<Skill[]>([
    { id: 'atlas',     title: 'Atlas Onboarding',     author: 'condura', version: '2.1.0', loaded: true,  trust: 'official',    description: 'Run a new user through your app in five steps.', tags: ['onboarding', 'tutorial'] },
    { id: 'onyx',      title: 'Onyx Governance',      author: 'sahaj',    version: '0.4.1', loaded: true,  trust: 'community',   description: 'Track who changed what in your repo, with line-level attribution.', tags: ['audit', 'git'] },
    { id: 'summarize', title: 'Daily Summarize',      author: 'condura', version: '1.2.0', loaded: true,  trust: 'official',    description: 'One-pager on what changed in your world today.', tags: ['summary', 'daily'] },
  ])

  let available = $state<Skill[]>([
    { id: 'atlas',     title: 'Atlas Onboarding',     author: 'condura', version: '2.1.0', loaded: true,  trust: 'official',    description: 'Run a new user through your app in five steps.', tags: ['onboarding', 'tutorial'] },
    { id: 'onyx',      title: 'Onyx Governance',      author: 'sahaj',    version: '0.4.1', loaded: true,  trust: 'community',   description: 'Track who changed what in your repo, with line-level attribution.', tags: ['audit', 'git'] },
    { id: 'summarize', title: 'Daily Summarize',      author: 'condura', version: '1.2.0', loaded: true,  trust: 'official',    description: 'One-pager on what changed in your world today.', tags: ['summary', 'daily'] },
    { id: 'travel',    title: 'Travel Planner',       author: 'maya',     version: '1.0.0', loaded: false, trust: 'community',   description: 'Draft an itinerary from your inbox + calendar.', tags: ['travel', 'inbox'] },
    { id: 'review',    title: 'PR Reviewer',          author: 'condura', version: '3.4.0', loaded: false, trust: 'official',    description: 'A second pair of eyes that knows your coding style.', tags: ['code', 'review'] },
    { id: 'morning',   title: 'Morning Briefing',     author: 'jordan',   version: '0.9.0', loaded: false, trust: 'experimental', description: 'Five-minute audio + a one-page summary at 7am.', tags: ['audio', 'daily'] },
    { id: 'notes',     title: 'Zettelkasten Notes',   author: 'kira',     version: '1.5.0', loaded: false, trust: 'community',   description: 'Auto-link your notes by concept, not just keyword.', tags: ['notes', 'graph'] },
    { id: 'finance',   title: 'Finance Watcher',      author: 'rishi',    version: '2.0.0', loaded: false, trust: 'community',   description: 'Track subscriptions and quietly flag what you stopped using.', tags: ['finance'] },
    { id: 'transit',   title: 'Transit Helper',       author: 'sahaj',    version: '1.1.0', loaded: false, trust: 'community',   description: 'Pre-arrival commute alerts tied to your calendar.', tags: ['travel', 'commute'] },
    { id: 'garden',    title: 'Garden Journal',       author: 'osman',    version: '0.3.0', loaded: false, trust: 'experimental', description: 'Weekly plant-check reminder tuned to your local weather.', tags: ['gardening'] },
    { id: 'meeting',   title: 'Meeting Minuteman',    author: 'condura', version: '1.7.0', loaded: false, trust: 'official',    description: 'Live transcripts + a one-page summary at the end.', tags: ['meetings', 'audio'] },
    { id: 'mail',      title: 'Inbox Triager',        author: 'priya',    version: '4.0.1', loaded: false, trust: 'community',   description: 'Sort the unread, draft the easy replies, flag the rest.', tags: ['email', 'inbox'] },
  ])
</script>

<div data-v2 style="
  min-height: 100vh;
  background: var(--v2-paper);
  display: flex;
  box-sizing: border-box;
  overflow: hidden;
">
  <Sidebar
    {items}
    {active}
    {collapsed}
    onSelect={(id) => active = id}
    onToggle={() => collapsed = !collapsed}
  />

  <div style="flex: 1; display: flex; flex-direction: column; min-height: 100vh; min-width: 0;">
    <StatusBar
      agentName="condura"
      currentTask={null}
      taskStartedAt={null}
      queueDepth={0}
      todaySpend="$0.0014"
      online={true}
      activeModel="ollama · qwen2.5-coder"
    />

    <Hub
      {query}
      onQueryChange={(q) => query = q}
      {installed}
      {available}
    />
  </div>
</div>
