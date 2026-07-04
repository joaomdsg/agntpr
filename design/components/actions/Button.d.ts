/**
 * Operational buttons — one accent action per screen, everything else bordered neutral.
 */
export interface ButtonProps {
  /** 'primary' (accent + glow, e.g. ＋ New packet) | 'secondary' (bordered neutral) | 'outline' (accent border) | 'pager' (◂ ▸) | 'cta' (marketing 50px) */
  variant?: 'primary' | 'secondary' | 'outline' | 'pager' | 'cta';
  disabled?: boolean;
  onClick?: () => void;
  title?: string;
  style?: React.CSSProperties;
  children?: React.ReactNode;
}
