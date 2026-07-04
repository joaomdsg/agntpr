/**
 * The locked app titlebar pattern (Inspector chrome).
 * @startingPoint section="Chrome" subtitle="App titlebar: lockup, status packet, identity block" viewport="700x120"
 */
export interface TitlebarProps {
  /** Product label in the stacked lockup (e.g. "inspector", "console"). */
  app?: string;
  /** Packet name — the bar's single white anchor. */
  name: string;
  /** Revision chip text (e.g. "rev2"). */
  rev?: string;
  /** The addr, below the name (owner/repo form, e.g. "acme/edge-gateway"). */
  addr?: string;
  /** Status packet: { state: 'risk'|'held'|'verified'|…, label: 'HELD · STRICT', detail: 'held 34m · …' }. Hover reveals the tooltip. */
  status?: { state?: string; label: string; detail?: string };
  /** Right-aligned content (annotation counts, links). */
  right?: React.ReactNode;
  style?: React.CSSProperties;
}
