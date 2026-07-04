/**
 * Pill-shaped mono microtype chip — the system's label unit.
 */
export interface ChipProps {
  /** 'neutral' (addr/rev/meta) or a state: 'signal' | 'you' | 'verified' | 'held' | 'risk' | 'agent' | 'delivered' */
  state?: string;
  /** 'sm' (15px) | 'md' (17px) | 'lg' (20px). Default 'md'. */
  size?: 'sm' | 'md' | 'lg';
  /** Add the 40%-mix 1px border (used on strong status chips). */
  bordered?: boolean;
  /** Leading 5px dot in the chip color (author chips, legends). */
  dot?: boolean;
  /** Neutral-chip selected look (cyan tint) — for active addr filters. */
  active?: boolean;
  /** Force casing; defaults to uppercase for state chips, none for neutral. */
  uppercase?: boolean;
  onClick?: () => void;
  title?: string;
  style?: React.CSSProperties;
  children?: React.ReactNode;
}
