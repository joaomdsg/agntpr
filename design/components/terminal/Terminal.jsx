import React from 'react';

/**
 * The CLI frame — three flat dots, a mono label, and a dark paper body.
 * Line color conventions: prompt/$ and comments --text-disabled,
 * ✓ --verified, ⌁ --signal, links --signal.
 */
export function Terminal({ title = 'ci · deploy.yml', style, children }) {
  return (
    <div style={{ border: '1px solid var(--hairline)', borderRadius: 'var(--r-frame)', background: '#0d1118', overflow: 'hidden', ...style }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 6, padding: '11px 14px', background: 'var(--surface-panel)', borderBottom: '1px solid var(--hairline)' }}>
        <span style={{ width: 9, height: 9, borderRadius: '50%', background: 'var(--hairline)' }} />
        <span style={{ width: 9, height: 9, borderRadius: '50%', background: 'var(--hairline)' }} />
        <span style={{ width: 9, height: 9, borderRadius: '50%', background: 'var(--hairline)' }} />
        <span style={{ marginLeft: 8, fontFamily: 'var(--font-mono)', fontSize: 10, color: 'var(--text-faint)' }}>{title}</span>
      </div>
      <div style={{ padding: '20px 22px', fontFamily: 'var(--font-mono)', fontSize: 12.5, lineHeight: 2, color: 'var(--text-muted)' }}>
        {children}
      </div>
    </div>
  );
}
