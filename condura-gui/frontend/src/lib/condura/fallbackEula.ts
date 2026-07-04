// Legal consent must happen before any system access, including daemon
// connectivity. This local copy of the EULA is the lawful source of consent
// when the daemon is unreachable; the recorded version is replayed to the
// daemon on next boot. The ritual is the only place this matters — do not
// use this fallback outside the first-run wizard.

export const FALLBACK_EULA_VERSION = 'v1';

export const FALLBACK_EULA_TEXT = `Condura (Synaptic) Freeware EULA v1

1. Grant of license. Condura is free for personal and commercial use. You may install and run it on any machine you own or control.

2. No redistribution. You may not redistribute the binary. Share the link to synaptic.app — not the file.

3. No tracking. Condura does not collect telemetry. It does not phone home. The only network calls are to the AI providers you configure.

4. Local-first. Your data — memory, skills, audit log, embeddings — lives on your machine, encrypted at rest. The only thing that leaves is what you send to your AI provider.

5. Revocable. We can revoke this license for abuse. Don't be abusive.

6. No warranty. Condura is provided "as is." It performs physical actions on your computer; you accept the risk by using it.

7. Your say-so. Every action that can't be undone requires your explicit consent. There is no "trust me." There is only your decision.

By stamping the seal below, you acknowledge that you have read and accept these terms.`;