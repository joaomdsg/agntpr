# Pakets — Tracking

## Locked decisions
- **Mark (2026-07-03):** 2×2 grid, variant A. TL/BL signal #4cc4d4, BR delivered #2a7683 solid, TR ghost (composing) — #2a7683 stroke at 18% of cell, 8% tint fill, scale 1.035. Gap 27% of cell, radius 23%. Small-size rule: below ~14px, ghost outline → solid #357f8a. Story: composed top-right → delivered bottom-right, same edge = same addr, same color.
- **Wordmark:** IBM Plex Mono 700 lowercase, +.04em tracking, lockup gap ≈ 1.2× cell.
- **Addr = repo** (owner/repo form, e.g. acme/edge-gateway). Never `addr/...`.
- **Inspector titlebar pattern:** logo mark → packets/INSPECTOR stacked lockup → hairline → packet name + rev chip, addr below → annotations right → status as single colored packet w/ hover tooltip.
- Spec lives in `Brand Board v2.dc.html` (locked spec card at bottom).

## Follow-ups — "tie it all together" concepts (from brand session)
1. **The mark is the console** ← preferred candidate. Logo as live 2×2 legend of real traffic: ghost = composing count, bright = in flight, dark = delivered, red cell = HELD. Logo in every screen corner is a tiny dashboard. Brand = telemetry.
2. **One packet, followed home.** Scroll-driven page following one packet: plain words → contract → flight → held → inspected → delivered; the same square changes state throughout — the logo revealed as four frames of one packet's life.
3. **The addr ledger.** Per-repo identity strip: packets woven as a row of squares (dark = delivered history, bright = in flight, outline = composing). Repo header / README badge / signature.
4. **Idle = compose loop.** Animated cycle (ghost fills → brightens → darkens → new ghost) as universal "system is alive" indicator: loading, empty states, favicon during CI.

## Lineage research (for a future "CD → Packets" page)
- **Dijkstra:** "testing shows presence, never absence of bugs" → why contracts = mutation testing + properties, not examples. EWD667 (natural language ≠ precision) → plain words are the *interface*, the handshake is the formal *artifact*. "Small heads" → the attention budget. THE-system layers → inspector's layer-by-layer decode.
- **Farley (Modern Software Engineering / CD):** pipeline = definitive releasability → handshake per change. Falsification mechanism → forward loop. Change Approval Boards negatively correlate w/ quality; capture human decisions in the pipeline → HELD as captured decision, not queue. Small batches → the packet as unit. Deploy≠release → delivered = ACK'd healthy in prod. "Branches tell lies" → 200 agent changes = 200 branches; packets re-verify against integrated state (collisions page). Pair-programming navigator → agent self-flags. <1h feedback → lane floors.
- Synthesis: Dijkstra "prove it" → Farley "pipeline it" → Packets "machines prove it; humans spend attention". Page idea: scroll timeline, each era a packet state (outline → bright → dark).

## Design system build (in progress)
- Done: old_stale/ archive, tokens/ + styles.css, guidelines/ cards + voice.md, readme.md (fundamentals/foundations/iconography), branding badges.
- Building: components/{cells,chips,actions,surfaces,annotations,chrome,timeline,terminal}, ui_kits/console (Console + Inspector), SKILL.md.
- Cleanup applied: DELIVERED chips use --delivered #2a7683 (Console used slate #98a1b2).

## Other open items
- Propagate locked mark recipe (colors/ratios/fallback) to: Landing Page nav + footer + inspector mock, Console, all other surfaces (currently on the old all-cyan recipe).
- Landing page inspector mock: HELD·STRICT only visible on hover — consider slow pulse or default-open tooltip (user said leave for now).
- Brand Board (v1) page: direction question asked, unanswered (brand book vs configurator).
- Older mark pages (Packets Mark 2x2, Pressure Test, Logo System, Lockups) predate the locked spec — reconcile or archive.
