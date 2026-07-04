import React from 'react';
import { PacketCell } from '../../components/cells/PacketCell.jsx';
import { PacketMark } from '../../components/cells/PacketMark.jsx';
import { Chip } from '../../components/chips/Chip.jsx';
import { Button } from '../../components/actions/Button.jsx';
import { Card } from '../../components/surfaces/Card.jsx';

const mono = (extra) => ({ fontFamily: 'var(--font-mono)', fontVariantNumeric: 'tabular-nums', ...extra });

function PanelHeader({ dot, children, right }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '12px 16px', borderBottom: '1px solid var(--hairline)', background: 'var(--surface-card)' }}>
      {dot && <PacketCell state={dot} size={8} />}
      <span style={mono({ fontSize: 9, fontWeight: 700, letterSpacing: 'var(--track-label)', color: 'var(--ink)', textTransform: 'uppercase', whiteSpace: 'nowrap' })}>{children}</span>
      {right && <span style={mono({ marginLeft: 'auto', fontSize: 9, color: 'var(--text-faint)', whiteSpace: 'nowrap' })}>{right}</span>}
    </div>
  );
}

const QUEUE = [
  { name: 'rate-limiter', net: 'acme/edge-gateway', tag: 'HELD · STRICT', state: 'held', why: 'Handshake is example-only on a strict lane · conf 0.72 · 1 mutant survived', age: 'held 34m', action: 'adjudicate' },
  { name: 'migrate-db', net: 'acme/billing', tag: 'IRREVERSIBLE', state: 'risk', why: 'Drops 12 legacy tables — mandatory human, verify-before-prod', age: 'held 2h', action: 'adjudicate' },
  { name: 'checkout-e2e', net: 'acme/checkout', tag: 'TRIGGER FIRE', state: 'agent', why: 'Your flaky-e2e watch fired — flake.count:4 in the same suite', age: '26m ago', action: 'open fire' },
];

const INFLIGHT = [
  { name: 'parse-retry rev4', net: 'acme/ingest', lane: 'standard', laneCol: 'var(--signal)', dot: 'verified', stage: 'verifying · property suite', prog: 72, t: '6m' },
  { name: 'ui-polish', net: 'acme/control-plane', lane: 'best-effort', laneCol: 'var(--verified)', dot: 'verified', stage: 'verifying · examples', prog: 81, t: '9m' },
  { name: 'dep-bump', net: 'acme/edge-gateway', lane: 'best-effort', laneCol: 'var(--verified)', dot: 'verified', stage: 'forwarding · building', prog: 38, t: '2m' },
  { name: 'retry-backoff', net: 'acme/ingest', lane: 'standard', laneCol: 'var(--signal)', dot: 'you', stage: 'composing · spec drafted', prog: 15, t: '40s' },
];

const LANES = [
  { name: 'best-effort', col: 'var(--verified)', vol: '96 pkts', spec: 'examples', trust: 92, trustCol: 'var(--verified)', note: 'all green · fully silent', noteCol: 'var(--text-faint)' },
  { name: 'standard', col: 'var(--signal)', vol: '87 pkts', spec: 'properties', trust: 71, trustCol: 'var(--signal)', note: 'floor raised after drift on tue — trust rebuilding', noteCol: 'var(--held)' },
  { name: 'strict', col: 'var(--risk)', vol: '28 pkts', spec: 'contracts', trust: 54, trustCol: 'var(--held)', note: '2 held for you right now', noteCol: 'var(--held)' },
  { name: 'irreversible', col: 'var(--risk-deep)', vol: '3 pkts', spec: 'contracts→proof', trust: 0, trustCol: 'var(--text-ghost)', note: 'never trusted · always verified before prod', noteCol: 'var(--text-faint)' },
];

const WATCHES = [
  { name: 'auth-mutants', dot: 'verified', meta: '86% signal', metaCol: 'var(--verified)' },
  { name: 'handshake-weakening', dot: 'verified', meta: '100% signal', metaCol: 'var(--verified)' },
  { name: 'flaky-e2e', dot: 'held', meta: '⚠ noisy', metaCol: 'var(--held)' },
];

const DELIVERED = [
  { name: 'cache-key-fix', net: 'acme/edge-gateway', t: '11m' },
  { name: 'docs-pass', net: 'acme/docs', t: '42m' },
  { name: 'parse-retry rev3', net: 'acme/ingest', t: '1h' },
];

const HOURS = [6, 7, 5, 4, 3, 2, 2, 3, 6, 10, 13, 11, 14, 12, 15, 12, 13, 16, 15, 12, 11, 9, 10, 3];
const HELD_AT = [10, 14, 17];

export function ConsoleScreen({ onOpenPacket }) {
  return (
    <div data-screen-label="Console" style={{ display: 'grid', gridTemplateRows: 'auto 1fr auto', height: '100%', background: 'var(--ground)', fontFamily: 'var(--font-ui)', fontSize: 13, color: 'var(--ink)', overflow: 'hidden' }}>
      {/* header */}
      <header style={{ display: 'flex', alignItems: 'center', gap: 16, padding: '12px 22px', background: 'var(--surface-card)', borderBottom: '1px solid var(--border-faint)' }}>
        <PacketMark cell={8} sub="console" />
        <span style={mono({ display: 'inline-flex', alignItems: 'center', gap: 6, fontSize: 10, color: 'var(--verified)' })}>
          <PacketCell state="verified" size={8} round live />stream live · 41/hr
        </span>
        <div style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: 20 }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4, alignItems: 'flex-end' }}>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
              <span style={mono({ fontSize: 9, fontWeight: 700, letterSpacing: '.08em', color: 'var(--text-faint)', textTransform: 'uppercase' })}>attention · this week</span>
              <span style={mono({ fontSize: 11, fontWeight: 700, color: 'var(--verified)' })}>5 / 10 interrupts</span>
            </div>
            <div style={{ width: 180, height: 6, borderRadius: 3, background: 'var(--surface-raised)', overflow: 'hidden' }}>
              <span style={{ display: 'block', height: '100%', width: '50%', background: 'var(--verified)' }} />
            </div>
          </div>
          <Button variant="primary" onClick={() => {}}>＋ New packet</Button>
        </div>
      </header>

      {/* body */}
      <div style={{ display: 'grid', gridTemplateColumns: '360px 1fr 340px', minHeight: 0, overflow: 'hidden' }}>
        {/* left · needs you */}
        <aside style={{ borderRight: '1px solid var(--hairline)', background: 'var(--surface-panel)', display: 'flex', flexDirection: 'column', minHeight: 0, overflow: 'auto' }}>
          <PanelHeader dot="held" right="the whole job">needs you · 4</PanelHeader>
          <div style={{ padding: 14, display: 'flex', flexDirection: 'column', gap: 11 }}>
            {QUEUE.map((q) => (
              <Card key={q.name} accent={`var(--${q.state === 'held' ? 'held' : q.state === 'risk' ? 'risk' : 'agent'})`} onClick={onOpenPacket}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <span style={mono({ fontSize: 11, fontWeight: 700, color: 'var(--ink)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' })}>{q.name}</span>
                  <span style={{ marginLeft: 'auto', flex: 'none' }}><Chip state={q.state} size="lg">{q.tag}</Chip></span>
                </div>
                <div style={mono({ fontSize: 9.5, color: 'var(--text-muted)', marginTop: 8, lineHeight: 1.6 })}>{q.why}</div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 10 }}>
                  <Chip size="md">{q.net}</Chip>
                  <span style={mono({ fontSize: 9.5, color: 'var(--text-faint)' })}>{q.age}</span>
                  <span style={mono({ marginLeft: 'auto', fontSize: 9, fontWeight: 700, color: 'var(--signal)' })}>{q.action} →</span>
                </div>
              </Card>
            ))}
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <Button variant="pager" disabled>◂</Button>
              <span style={mono({ flex: 1, textAlign: 'center', fontSize: 9, color: 'var(--text-faint)' })}>seq 01–03 / 04 · window 3</span>
              <Button variant="pager">▸</Button>
            </div>
            <Card variant="dashed">
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <span style={mono({ fontSize: 10, fontWeight: 700, color: 'var(--signal)' })}>calibration draw</span>
                <span style={mono({ marginLeft: 'auto', fontSize: 9, color: 'var(--text-faint)' })}>3 of 214 · random</span>
              </div>
              <div style={mono({ fontSize: 9.5, color: 'var(--text-muted)', marginTop: 8, lineHeight: 1.6 })}>Skim three auto-merged packets to keep your trust measured, not assumed. ~6 min.</div>
              <div style={{ marginTop: 11 }}><Button variant="outline">Skim the sample</Button></div>
            </Card>
            <div style={{ display: 'flex', gap: 8, fontSize: 12, lineHeight: 1.55, color: 'var(--signal)', padding: '4px 5px 10px' }}>
              <span>✱</span>
              <span>Everything not on this list forwarded without you — that's the point. An empty queue is success, not idleness.</span>
            </div>
          </div>
        </aside>

        {/* center · the stream */}
        <main style={{ display: 'flex', flexDirection: 'column', minWidth: 0, minHeight: 0, overflow: 'auto', background: 'var(--ground)', backgroundImage: 'var(--wash-signal)' }}>
          <div style={{ padding: '30px 32px 0' }}>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 16, flexWrap: 'wrap' }}>
              <span style={mono({ fontSize: 'var(--fs-hero-stat)', fontWeight: 600, letterSpacing: '-.02em', color: 'var(--ink)', lineHeight: 1 })}>214</span>
              <span style={{ fontSize: 12.5, color: 'var(--text-muted)' }}>packets forwarded without you today</span>
              <span style={mono({ marginLeft: 'auto', fontSize: 10, color: 'var(--text-faint)' })}>last 24h</span>
            </div>
            <div style={{ display: 'flex', gap: 18, marginTop: 14, flexWrap: 'wrap' }}>
              {[['verified', '153 verified'], ['delivered', '53 delivered'], ['held', '7 sampled & held'], ['risk', '1 dropped'], ['you', '8 in flight']].map(([s, t]) => (
                <span key={s} style={mono({ display: 'inline-flex', alignItems: 'center', gap: 6, fontSize: 10, color: `var(--${s === 'you' ? 'signal' : s})` })}>
                  <PacketCell state={s} size={8} />{t}
                </span>
              ))}
            </div>
            <div style={{ display: 'flex', alignItems: 'flex-end', gap: 3, height: 72, marginTop: 24, borderBottom: '1px solid var(--hairline)', paddingBottom: 1 }}>
              {HOURS.map((n, i) => (
                <div key={i} style={{ flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'flex-end', gap: 1 }}>
                  {HELD_AT.includes(i) && <span style={{ display: 'block', width: '100%', height: 4, borderRadius: 2, background: 'var(--held)' }} />}
                  <span style={{ display: 'block', width: '100%', borderRadius: '2px 2px 0 0', height: 6 + n * 4.2, background: 'var(--verified)' }} />
                </div>
              ))}
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 6 }}>
              <span style={mono({ fontSize: 9, color: 'var(--text-ghost)' })}>-24h</span>
              <span style={mono({ fontSize: 9, color: 'var(--text-faint)' })}><PacketCell state="held" size={7} style={{ verticalAlign: 'middle', marginRight: 5 }} />tick = a packet held that hour</span>
              <span style={mono({ fontSize: 9, color: 'var(--text-ghost)' })}>now</span>
            </div>
          </div>

          <div style={{ padding: '26px 32px 0' }}>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 10 }}>
              <span style={mono({ fontSize: 9, fontWeight: 700, letterSpacing: 'var(--track-label)', color: 'var(--text-faint)', textTransform: 'uppercase' })}>in flight · 8</span>
              <span style={mono({ marginLeft: 'auto', fontSize: 9, color: 'var(--text-ghost)' })}>the machine at work — nothing to do here</span>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginTop: 12 }}>
              {INFLIGHT.map((p) => (
                <Card key={p.name} variant="row" style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  <PacketCell state={p.dot} size={7} round live />
                  <span style={mono({ fontSize: 10.5, fontWeight: 700, color: 'var(--ink)', whiteSpace: 'nowrap', flex: 'none' })}>{p.name}</span>
                  <Chip size="md">{p.net}</Chip>
                  <span style={mono({ fontSize: 9, color: p.laneCol })}>{p.lane}</span>
                  <span style={mono({ marginLeft: 'auto', fontSize: 9.5, color: 'var(--text-muted)' })}>{p.stage}</span>
                  <span style={{ display: 'inline-flex', width: 72, height: 5, borderRadius: 3, background: 'var(--surface-raised)', overflow: 'hidden' }}>
                    <span style={{ display: 'block', height: '100%', width: p.prog + '%', borderRadius: 3, background: 'color-mix(in srgb, var(--verified) 55%, var(--surface-raised))' }} />
                  </span>
                  <span style={mono({ fontSize: 9, color: 'var(--text-ghost)', width: 30, textAlign: 'right' })}>{p.t}</span>
                </Card>
              ))}
            </div>
          </div>

          <div style={{ padding: '28px 32px 32px' }}>
            <span style={mono({ fontSize: 9, fontWeight: 700, letterSpacing: 'var(--track-label)', color: 'var(--text-faint)', textTransform: 'uppercase' })}>lane health</span>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 12, marginTop: 14 }}>
              {LANES.map((l) => (
                <Card key={l.name}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <span style={{ width: 9, height: 9, borderRadius: '50%', background: l.col }} />
                    <span style={mono({ fontSize: 10.5, fontWeight: 700, color: 'var(--ink)' })}>{l.name}</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 13 }}>
                    <span style={mono({ fontSize: 9, color: 'var(--text-faint)' })}>this week</span>
                    <span style={mono({ fontSize: 9, fontWeight: 700, color: 'var(--ink)' })}>{l.vol}</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 7 }}>
                    <span style={mono({ fontSize: 9, color: 'var(--text-faint)' })}>term floor</span>
                    <span style={mono({ fontSize: 9, fontWeight: 700, color: l.col })}>{l.spec}</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 7 }}>
                    <span style={mono({ fontSize: 9, color: 'var(--text-faint)' })}>trust</span>
                    <span style={{ display: 'inline-flex', width: 64, height: 6, borderRadius: 3, background: 'var(--surface-raised)', overflow: 'hidden' }}>
                      <span style={{ display: 'block', height: '100%', width: l.trust + '%', background: l.trustCol }} />
                    </span>
                  </div>
                  <div style={mono({ fontSize: 9.5, color: l.noteCol, marginTop: 12, lineHeight: 1.55 })}>{l.note}</div>
                </Card>
              ))}
            </div>
          </div>
        </main>

        {/* right · watches + delivered */}
        <aside style={{ borderLeft: '1px solid var(--hairline)', background: 'var(--surface-panel)', display: 'flex', flexDirection: 'column', minHeight: 0, overflow: 'auto' }}>
          <PanelHeader right="control room →">your watches</PanelHeader>
          <div style={{ padding: '13px 14px', display: 'flex', flexDirection: 'column', gap: 8 }}>
            {WATCHES.map((w) => (
              <Card key={w.name} padding="10px 12px" style={{ display: 'flex', alignItems: 'center', gap: 9, borderRadius: 'var(--r-card-sm)' }}>
                <PacketCell state={w.dot} size={7} round />
                <span style={mono({ fontSize: 10, fontWeight: 600, color: 'var(--ink)' })}>{w.name}</span>
                <span style={mono({ marginLeft: 'auto', fontSize: 9.5, color: w.metaCol })}>{w.meta}</span>
              </Card>
            ))}
          </div>
          <PanelHeader>friday digest · 5 collected</PanelHeader>
          <div style={{ padding: '13px 14px', display: 'flex', flexDirection: 'column', gap: 7 }}>
            {[['3×', 'new external deps landed on standard lane'], ['1×', 'term floor raised: standard → properties'], ['1×', 'trigger demoted: flaky-e2e → digest']].map(([n, t]) => (
              <div key={t} style={{ display: 'flex', alignItems: 'center', gap: 9, padding: '5px 3px' }}>
                <span style={mono({ fontSize: 9, color: 'var(--text-faint)', flex: 'none' })}>{n}</span>
                <span style={mono({ fontSize: 9.5, color: 'var(--text-muted)', lineHeight: 1.5 })}>{t}</span>
              </div>
            ))}
            <div style={mono({ fontSize: 9.5, color: 'var(--text-ghost)', padding: '6px 3px 2px' })}>lands friday 09:00 · nothing here can interrupt you</div>
          </div>
          <PanelHeader>recently delivered · 12</PanelHeader>
          <div style={{ padding: '13px 14px', display: 'flex', flexDirection: 'column', gap: 8 }}>
            {DELIVERED.map((d) => (
              <Card key={d.name} padding="9px 12px" style={{ display: 'flex', flexDirection: 'column', gap: 7, borderRadius: 'var(--r-card-sm)' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 9 }}>
                  <span style={mono({ fontSize: 10, fontWeight: 600, color: 'var(--ink)' })}>{d.name}</span>
                  <span style={mono({ marginLeft: 'auto', fontSize: 9.5, color: 'var(--text-faint)' })}>{d.t}</span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 9 }}>
                  <Chip size="sm">{d.net}</Chip>
                  <span style={{ marginLeft: 'auto' }}><Chip state="delivered" size="sm">DELIVERED</Chip></span>
                </div>
              </Card>
            ))}
          </div>
        </aside>
      </div>

      {/* footer */}
      <footer style={{ height: 36, display: 'flex', alignItems: 'center', gap: 16, padding: '0 18px', background: 'var(--surface-deep)', color: 'var(--text-muted)', borderTop: '1px solid var(--hairline)' }}>
        <span style={mono({ display: 'inline-flex', alignItems: 'center', gap: 6, fontSize: 9, color: 'var(--verified)', fontWeight: 600 })}>
          <PacketCell state="verified" size={7} round live />forwarding at line rate
        </span>
        <span style={mono({ fontSize: 9, color: 'var(--text-faint)' })}>4 held · 8 in flight · queue healthy</span>
        <span style={mono({ marginLeft: 'auto', fontSize: 9, color: 'var(--text-faint)' })}>⌘K filter the stream · same language everywhere</span>
      </footer>
    </div>
  );
}
