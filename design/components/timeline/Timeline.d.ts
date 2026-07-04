/**
 * Horizontal packet-life event track.
 */
export interface TimelineProps {
  /** Ordered events: { id?, t: '09:02'|'now', label, type: 'plan'|'edit'|'test'|'mutation'|'comment'|'flag'|'held'|'verified'|'delivered', rev?, big? } */
  events: Array<{ id?: string; t: string; label: string; type: string; rev?: string; big?: boolean }>;
  /** Selected event id (or index) — gets a ring + bold label. */
  selected?: string | number;
  /** Makes events clickable. */
  onSelect?: (id: string | number) => void;
  style?: React.CSSProperties;
}
