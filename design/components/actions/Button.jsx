import React from 'react';

/**
 * Operational buttons. primary = the one accent action per screen;
 * secondary = bordered neutral; outline = accent-bordered emphasis;
 * pager = ◂ ▸ seq controls; cta = marketing only.
 */
export function Button({ variant = 'secondary', disabled = false, onClick, title, style, children }) {
  const V = {
    primary: {
      height: 34, padding: '0 16px', borderRadius: 'var(--r-btn)', border: 'none',
      background: 'var(--signal)', color: 'var(--on-signal)', fontSize: 11, fontWeight: 700,
      boxShadow: 'var(--glow-btn)',
    },
    secondary: {
      height: 26, padding: '0 12px', borderRadius: 'var(--r-btn-sm)', border: '1px solid rgba(255,255,255,.12)',
      background: 'var(--surface-card)', color: 'var(--ink)', fontSize: 10, fontWeight: 600,
    },
    outline: {
      height: 27, padding: '0 13px', borderRadius: 'var(--r-btn-sm)', border: '1.5px solid var(--signal)',
      background: 'transparent', color: 'var(--signal)', fontSize: 10.5, fontWeight: 700,
    },
    pager: {
      height: 22, padding: '0 9px', borderRadius: 6, border: '1px solid rgba(255,255,255,.08)',
      background: 'var(--surface-card)', color: disabled ? 'var(--text-disabled)' : 'var(--text-muted)', fontSize: 10, fontWeight: 400,
    },
    cta: {
      height: 50, padding: '0 26px', borderRadius: 'var(--r-cta)', border: 'none',
      background: 'var(--signal)', color: '#0a0d13', fontSize: 14, fontWeight: 700,
      boxShadow: 'var(--glow-cta)',
    },
  }[variant];

  return (
    <button
      onClick={disabled ? undefined : onClick}
      title={title}
      style={{
        display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: 7,
        fontFamily: 'var(--font-mono)', whiteSpace: 'nowrap',
        cursor: disabled ? 'default' : 'pointer', opacity: disabled && variant !== 'pager' ? 0.45 : 1,
        ...V, ...style,
      }}
    >
      {children}
    </button>
  );
}
