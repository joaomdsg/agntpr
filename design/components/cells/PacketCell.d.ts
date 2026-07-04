/**
 * One packet, one square. The atomic state marker used in legends,
 * headers, chips and the mark itself.
 */
export interface PacketCellProps {
  /** 'composing' | 'inflight' | 'you' | 'verified' | 'held' | 'risk' | 'delivered' | 'agent' — or any CSS color */
  state?: string;
  /** Side length in px. Default 8. Composing below 14px auto-falls back to solid --delivered-mid. */
  size?: number;
  /** Live things pulse (risk pulses its red glow). Default false. */
  live?: boolean;
  /** Render as a dot (50% radius) instead of a rounded square. Default false. */
  round?: boolean;
  style?: React.CSSProperties;
}
