import React, { useState } from 'react';
import { PacketCell } from '../../components/cells/PacketCell.jsx';
import { Chip } from '../../components/chips/Chip.jsx';
import { Button } from '../../components/actions/Button.jsx';
import { AnnotationCard } from '../../components/annotations/AnnotationCard.jsx';
import { Titlebar } from '../../components/chrome/Titlebar.jsx';
import { Timeline } from '../../components/timeline/Timeline.jsx';

const mono = (extra) => ({ fontFamily: 'var(--font-mono)', fontVariantNumeric: 'tabular-nums', ...extra });

const FILES = [
  { name: 'limiter.go', dir: 'pkg/ratelimit', status: 'M', add: 3, del: 2, sel: true, badges: [{ state: 'agent', n: '2 ⚑' }] },
  { name: 'handler.go', dir: 'pkg/auth', status: 'M', add: 12, del: 4, badges: [{ state: 'you', n: '1' }] },
  { name: 'limiter_test.go', dir: 'pkg/ratelimit', status: 'A', add: 48, del: 0, badges: [{ state: 'you', n: '1' }] },
  { name: 'rate_limit.feature', dir: 'handshake', status: 'A', add: 22, del: 0, badges: [] },
];

const EVENTS = [
  { id: 'e1', t: '09:02', type: 'plan', label: 'plan approved', rev: 'rev0.5' },
  { id: 'e2', t: '09:14', type: 'edit', label: 'implement limiter', rev: 'rev1' },
  { id: 'e3', t: '09:15', type: 'test', label: 'tests written', rev: 'rev1' },
  { id: 'e4', t: '09:21', type: 'mutation', label: '1 mutant survived', rev: 'rev1' },
  { id: 'e5', t: '09:30', type: 'comment', label: 'you: reset property', rev: 'rev1' },
  { id: 'e6', t: '09:42', type: 'edit', label: 'revise · rev2', rev: 'rev2' },
  { id: 'e7', t: '09:50', type: 'flag', label: 'agent self-flag', rev: 'rev2' },
  { id: 'e8', t: 'now', type: 'held', label: 'held for review' },
];

function DiffLine({ n, kind, children, right }) {
  const bg = kind === 'add' ? 'color-mix(in srgb, var(--agent) 10%, var(--surface-card))' : kind === 'del' ? 'color-mix(in srgb, var(--risk) 8%, var(--surface-card))' : 'transparent';
  return (
    <div style={{ display: 'flex', paddingRight: 12, background: bg, borderLeft: kind === 'add' ? '3px solid var(--agent)' : '3px solid transparent', color: kind === 'add' ? 'var(--ink)' : kind === 'del' ? 'var(--risk-muted)' : 'var(--text-muted)' }}>
      <span style={{ width: 39, flex: 'none', textAlign: 'right', color: 'var(--text-disabled)', paddingRight: 12 }}>{n}</span>
      <span style={{ width: 16, flex: 'none', color: kind === 'add' ? 'var(--add)' : kind === 'del' ? 'var(--del)' : 'inherit', fontWeight: 700 }}>{kind === 'add' ? '+' : kind === 'del' ? '−' : ''}</span>
      <span style={{ whiteSpace: 'pre', flex: 1, overflow: 'hidden', textOverflow: 'ellipsis' }}>{children}</span>
      {right}
    </div>
  );
}

export function InspectorScreen({ onBack }) {
  const [sel, setSel] = useState('e6');
  const K = ({ children }) => <span style={{ color: 'var(--agent)' }}>{children}</span>;
  const T = ({ children }) => <span style={{ color: 'var(--signal)' }}>{children}</span>;
  const F = ({ children }) => <span style={{ color: 'var(--held)' }}>{children}</span>;

  return (
    <div data-screen-label="Inspector" style={{ display: 'grid', gridTemplateRows: 'auto 1fr auto', height: '100%', background: 'var(--ground)', fontFamily: 'var(--font-ui)', color: 'var(--ink)', overflow: 'hidden' }}>
      <Titlebar
        app="inspector" name="rate-limiter" rev="rev2" addr="acme/edge-gateway"
        status={{ state: 'risk', label: 'HELD · STRICT', detail: 'held 34m · handshake below lane floor' }}
        right={
          <>
            {onBack && <span onClick={onBack} style={mono({ fontSize: 10, fontWeight: 700, color: 'var(--signal)', cursor: 'pointer' })}>← console</span>}
            <span style={mono({ fontSize: 10, fontWeight: 600, color: 'var(--signal)' })}>✎ you · 2</span>
            <span style={mono({ fontSize: 10, fontWeight: 600, color: 'var(--agent)' })}>⚑ agents · 2</span>
          </>
        }
      />

      <div style={{ display: 'grid', gridTemplateColumns: '252px 1fr 312px', minHeight: 0, overflow: 'hidden' }}>
        {/* file tree */}
        <aside style={{ borderRight: '1px solid var(--hairline)', background: 'var(--surface-panel)', display: 'flex', flexDirection: 'column', minHeight: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 7, padding: '9px 12px', borderBottom: '1px solid var(--hairline)' }}>
            <span style={mono({ fontSize: 9, fontWeight: 700, letterSpacing: 'var(--track-label)', color: 'var(--text-muted)', textTransform: 'uppercase' })}>changed files</span>
            <span style={mono({ marginLeft: 'auto', fontSize: 9, fontWeight: 700, color: 'var(--text-faint)' })}>4</span>
          </div>
          <div style={{ flex: 1, overflow: 'auto', padding: '7px 7px 12px' }}>
            {FILES.map((f) => (
              <div key={f.name} style={{ border: '1px solid ' + (f.sel ? 'color-mix(in srgb, var(--signal) 45%, var(--surface-card))' : 'var(--hairline)'), background: f.sel ? 'color-mix(in srgb, var(--signal) 10%, var(--surface-card))' : 'var(--surface-card)', borderRadius: 'var(--r-card-sm)', padding: '8px 9px', marginBottom: 6, cursor: 'pointer' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 7 }}>
                  <span style={mono({ width: 15, height: 15, borderRadius: 'var(--r-glyph)', flex: 'none', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 9, fontWeight: 700, background: `color-mix(in srgb, var(--${f.status === 'A' ? 'verified' : 'held'}) 20%, var(--surface-card))`, color: `var(--${f.status === 'A' ? 'verified' : 'held'})` })}>{f.status}</span>
                  <span style={mono({ fontSize: 11.5, fontWeight: 600, color: 'var(--ink)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' })}>{f.name}</span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 7, marginTop: 6, paddingLeft: 22 }}>
                  <span style={mono({ fontSize: 9.5, color: 'var(--add)' })}>+{f.add}</span>
                  <span style={mono({ fontSize: 9.5, color: 'var(--del)' })}>−{f.del}</span>
                  <span style={mono({ fontSize: 9, color: 'var(--text-faint)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' })}>{f.dir}</span>
                  <span style={{ marginLeft: 'auto', display: 'flex', gap: 4 }}>
                    {f.badges.map((b, i) => <Chip key={i} state={b.state} size="sm" dot uppercase={false}>{b.n}</Chip>)}
                  </span>
                </div>
              </div>
            ))}
            <div style={{ marginTop: 4, padding: '8px 9px', border: '1px dashed color-mix(in srgb, var(--risk) 40%, var(--surface-card))', borderRadius: 'var(--r-card-sm)', background: 'color-mix(in srgb, var(--risk) 10%, var(--surface-card))' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 7 }}>
                <span style={mono({ width: 15, height: 15, borderRadius: 'var(--r-glyph)', flex: 'none', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 9, fontWeight: 700, background: 'color-mix(in srgb, var(--risk) 22%, var(--surface-card))', color: 'var(--risk)' })}>⚑</span>
                <span style={mono({ fontSize: 10.5, fontWeight: 600, color: 'var(--risk)' })}>go.mod</span>
              </div>
              <div style={mono({ fontSize: 9, color: 'var(--risk-muted)', marginTop: 5, paddingLeft: 22 })}>guardrail · must not change</div>
            </div>
          </div>
        </aside>

        {/* diff */}
        <main style={{ display: 'flex', flexDirection: 'column', minWidth: 0, minHeight: 0, background: 'var(--surface-card)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '0 12px', height: 36, background: 'var(--surface-panel)', borderBottom: '1px solid var(--hairline)', flex: 'none' }}>
            <span style={mono({ fontSize: 11, fontWeight: 600, color: 'var(--ink)' })}>pkg/ratelimit/limiter.go</span>
            <span style={mono({ fontSize: 9.5, color: 'var(--text-faint)' })}>func Allow · L13–27</span>
            <div style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: 7 }}>
              <Button variant="outline" style={{ height: 28, fontSize: 11 }}>✎ Annotate selection</Button>
              <Button style={{ height: 28, fontSize: 11 }}>⇄ split</Button>
            </div>
          </div>
          <div className="mono" style={{ flex: 1, minHeight: 0, overflow: 'auto', fontSize: 12, lineHeight: 1.9, padding: '10px 0 4px', fontFamily: 'var(--font-mono)' }}>
            <DiffLine n="13"><K>func</K> (l *<T>Limiter</T>) Allow(ip <T>string</T>) <T>bool</T> {'{'}</DiffLine>
            <DiffLine n="14">{'    now := time.Now()'}</DiffLine>
            <DiffLine n="18">{'    '}<K>for</K>{' _, t := '}<K>range</K>{' l.hits[ip] {'}</DiffLine>
            <DiffLine n="20">{'        recent = '}<F>append</F>{'(recent, t)'}</DiffLine>
            <DiffLine n="22">{'    }'}</DiffLine>
            <DiffLine n="23" kind="add" right={<span style={mono({ color: 'var(--agent)', fontSize: 9.5, paddingRight: 4, alignSelf: 'center' })}>agent-S · rev2</span>}>{'    l.hits[ip] = recent'}</DiffLine>
            <DiffLine n="24" kind="del">{'    '}<K>if</K>{' len(recent) > l.max {'}</DiffLine>
            <DiffLine n="25" kind="add" right={<span style={{ alignSelf: 'center', marginRight: 4 }}><Chip state="agent" size="sm" uppercase={false}>⚑ self-flag</Chip></span>}>{'    '}<K>if</K>{' len(recent) >= l.max {'}</DiffLine>
            <div style={{ margin: '6px 14px 8px 58px', maxWidth: 520 }}>
              <AnnotationCard author="agent" sev="self-flag" where="L25 · 09:50" actions="reply · resolve · promote to handshake term →">
                Unsure about this boundary — inclusive (&gt;=) or exclusive (&gt;)? The handshake only gives examples.
              </AnnotationCard>
            </div>
            <DiffLine n="26">{'        '}<K>return</K> <F>false</F></DiffLine>
            <DiffLine n="27">{'    }'}</DiffLine>
          </div>
          <div style={{ height: 25, flex: 'none', display: 'flex', alignItems: 'center', gap: 13, padding: '0 13px', background: 'var(--surface-deep)', whiteSpace: 'nowrap', overflow: 'hidden' }}>
            <span style={mono({ fontSize: 9, color: 'var(--signal)', fontWeight: 700 })}>● strict</span>
            <span style={mono({ fontSize: 9, color: 'var(--text-faint)' })}>conf 0.72</span>
            <span style={mono({ fontSize: 9, color: 'var(--text-faint)' })}>4 files · +85 −6</span>
            <span style={mono({ fontSize: 9, color: 'var(--agent)' })}>1 mutant survived rev1 · killed rev2</span>
            <span style={mono({ marginLeft: 'auto', fontSize: 9, color: 'var(--verified)' })}>⌥click a line to annotate · drag to group</span>
          </div>
        </main>

        {/* annotation rail */}
        <aside style={{ borderLeft: '1px solid var(--hairline)', background: 'var(--surface-panel)', display: 'flex', flexDirection: 'column', minHeight: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 7, padding: '9px 12px', borderBottom: '1px solid var(--hairline)' }}>
            <span style={mono({ fontSize: 9, fontWeight: 700, letterSpacing: 'var(--track-label)', color: 'var(--text-muted)', textTransform: 'uppercase' })}>annotations · this file</span>
            <span style={mono({ marginLeft: 'auto', fontSize: 9, fontWeight: 700, color: 'var(--text-faint)' })}>2</span>
          </div>
          <div style={{ flex: 1, overflow: 'auto', padding: '11px 12px', display: 'flex', flexDirection: 'column', gap: 10 }}>
            <AnnotationCard author="agent" sev="self-flag" where="L25" scope="line · limiter.go">
              Unsure about this boundary — should the limit be inclusive (&gt;=) or exclusive (&gt;)? The handshake only gives examples.
            </AnnotationCard>
            <AnnotationCard author="you" sev="blocking" where="L13–32" scope="group · limiter.go">
              This whole Allow() path needs a property test for the reset window, not just the 7th-request example.
            </AnnotationCard>
            <div style={{ marginTop: 'auto', border: '1px dashed var(--border-dashed)', borderRadius: 'var(--r-ann)', padding: '9px 10px', textAlign: 'center' }}>
              <span style={mono({ fontSize: 9.5, color: 'var(--text-faint)' })}>the agent revises · you never rubber-stamp</span>
            </div>
          </div>
        </aside>
      </div>

      {/* timeline */}
      <footer style={{ background: 'var(--surface-panel)', borderTop: '1px solid var(--hairline)', padding: '8px 14px 10px' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, paddingBottom: 4 }}>
          <span style={mono({ fontSize: 9, fontWeight: 700, letterSpacing: 'var(--track-label)', color: 'var(--text-muted)', textTransform: 'uppercase' })}>timeline · all actions</span>
          <div style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: 11 }}>
            {[['you', 'you'], ['agent', 'agents'], ['held', 'tests'], ['risk', 'held']].map(([s, n]) => (
              <span key={n} style={mono({ display: 'inline-flex', alignItems: 'center', gap: 5, fontSize: 9.5, fontWeight: 600, color: 'var(--text-muted)' })}>
                <PacketCell state={s} size={8} />{n}
              </span>
            ))}
          </div>
        </div>
        <Timeline events={EVENTS} selected={sel} onSelect={setSel} />
      </footer>
    </div>
  );
}
