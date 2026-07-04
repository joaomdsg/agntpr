---
name: packets-design
description: Use this skill to generate well-branded interfaces and assets for Packets (the agent-code control plane), either for production or throwaway prototypes/mocks/etc. Contains essential design guidelines, colors, type, fonts, assets, and UI kit components for prototyping.
user-invocable: true
---

Read the readme.md file within this skill, and explore the other available files. `guidelines/concepts.md` is the product's conceptual model (packet, addr, handshake, lanes, the gauntlet, inspection modes, attention economics) — the vocabulary there is binding for both copy and UI. `guidelines/lineage.md` carries the Dijkstra → Farley → Packets intellectual lineage for narrative/marketing work.

Non-negotiables when designing for Packets: the state grammar is a color system (composing = outline, in-flight = cyan, verified = green, held = amber, blocking = red, agent = purple, delivered = dark cyan) — never invent a new state color. Mono (IBM Plex Mono) is the machine's voice; sans only for prose. The 2×2 mark is constructed in code (components/cells/PacketMark), never drawn or rasterized. Icons are unicode glyphs, never icon libraries or hand-drawn SVG. No exclamation marks, no emoji, lowercase operational copy, networking vocabulary (forward/hold/inspect/deliver/ACK/addr/lane — never PR/merge/approve).

If creating visual artifacts (slides, mocks, throwaway prototypes, etc), copy assets out and create static HTML files for the user to view. If working on production code, you can copy assets and read the rules here to become an expert in designing with this brand.

If the user invokes this skill without any other guidance, ask them what they want to build or design, ask some questions, and act as an expert designer who outputs HTML artifacts _or_ production code, depending on the need.
