/**
 * Annotation card — the conversation between you and the agents,
 * pinned to code. Authorship = 3px left border.
 */
export interface AnnotationCardProps {
  /** 'you' (cyan) | 'agent' (purple). Default 'you'. */
  author?: 'you' | 'agent';
  /** 'blocking' (risk) | 'self-flag' (agent) | 'question' (held) | 'nit' (neutral). Default 'nit'. */
  sev?: string;
  /** Location meta, right-aligned in the header (e.g. "L25", "L13–32"). */
  where?: string;
  /** Scope footnote line (e.g. "line · limiter.go"). */
  scope?: string;
  /** Action links line (e.g. "reply · resolve · promote to handshake term →"). */
  actions?: React.ReactNode;
  onClick?: () => void;
  style?: React.CSSProperties;
  /** The annotation body — sans prose, the machine's mono stops here. */
  children?: React.ReactNode;
}
