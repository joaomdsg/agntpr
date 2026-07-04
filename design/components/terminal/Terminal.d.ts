/**
 * CLI frame for packets commands and CI output.
 */
export interface TerminalProps {
  /** Header label (e.g. "ci · deploy.yml"). */
  title?: string;
  style?: React.CSSProperties;
  /** Mono lines. Color spans inline: $ / comments --text-disabled, ✓ --verified, ⌁ --signal. */
  children?: React.ReactNode;
}
