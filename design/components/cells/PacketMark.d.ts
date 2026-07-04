/**
 * The locked 2×2 brand mark (built in code — no raster asset, deliberately).
 * @startingPoint section="Brand" subtitle="The locked packets mark + lockups" viewport="700x180"
 */
export interface PacketMarkProps {
  /** Cell size in px. Default 8 (app chrome). Nav uses 8, brand surfaces 22+. */
  cell?: number;
  /** Render mark + "packets" wordmark lockup. Default false (mark only). */
  wordmark?: boolean;
  /** Override wordmark font-size (default ≈ 1.7 × cell). */
  wordmarkSize?: number;
  /** Stacked product lockup: "packets" over this uppercase label (e.g. "inspector"). */
  sub?: string;
  /** Live-state variant: bottom-left cell burns red and pulses (something is held). */
  held?: boolean;
  style?: React.CSSProperties;
}
