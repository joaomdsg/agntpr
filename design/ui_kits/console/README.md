# UI kit — Packets product app

Recreations of the two core product surfaces, composed from `components/`:

- **ConsoleScreen.jsx** — the operator home (1440×900): header with attention budget + ＋ New packet, "needs you" queue (left), the stream (hero count, state legend, 24h throughput bars with held ticks, in-flight rows, lane health), watches / friday digest / recently delivered (right), live status footer.
- **InspectorScreen.jsx** — the capture-file decode of one held packet: locked titlebar (status packet + identity block), changed-files tree with guardrail card, authorship-marked diff with inline agent self-flag annotation, annotation rail, strict statusbar, replayable timeline.
- **index.html** — interactive: tab between screens; clicking a queue card in the Console opens the Inspector.

Deliberate deviations from the plain-networking baseline (agreed cleanups):
- DELIVERED chips use `--delivered` #2a7683 (Console used slate).
- Inspector titlebar uses the locked lockup pattern, not the old breadcrumb.
- Diff pane is a faithful static rendition (the original embeds Monaco; a kit doesn't need a live editor).
- Lists are abbreviated (3–4 rows standing in for the full seq windows).
