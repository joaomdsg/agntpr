/**
 * Bounded surface — 1px hairline, flat fill, no shadow.
 */
export interface CardProps {
  /** 'card' (default, --surface-card) | 'row' (panel-toned list row) | 'dashed' (empty/aside state) | 'tile' (raised, dense) */
  variant?: 'card' | 'row' | 'dashed' | 'tile';
  /** 3px left border color — authorship/queue accent. Use state colors only. */
  accent?: string;
  /** Override the variant's default padding. */
  padding?: string;
  onClick?: () => void;
  style?: React.CSSProperties;
  children?: React.ReactNode;
}
