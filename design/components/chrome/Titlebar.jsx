import React, { useState } from 'react';
import { PacketMark } from '../cells/PacketMark.jsx';
import { PacketCell } from '../cells/PacketCell.jsx';
import { Chip } from '../chips/Chip.jsx';

/**
 * The locked app titlebar: mark → stacked lockup → hairline → status
 * packet (hover for detail) → packet identity block; annotations
 * legend right. Status is a single colored packet — one packet, one square.
 */
export function Titlebar({ app = 'inspector', name, rev, addr, status, right, style }) {
  const [tip, setTip] = useState(false);
  return (
    <header
      style={{
        display: 'flex', alignItems: 'center', gap: 14, padding: '10px 16px',
        background: 'var(--surface-panel)', borderBottom: '1px solid var(--hairline)',
        ...style,
      }}
    >
      <PacketMark cell={8} sub={app} />
      <span style={{ width: 1, height: 26, background: 'var(--hairline)', flex: 'none' }} />
      {status && (
        <span
          onMouseEnter={() => setTip(true)}
          onMouseLeave={() => setTip(false)}
          style={{ position: 'relative', display: 'inline-flex', flex: 'none', cursor: 'default' }}
        >
          <PacketCell state={status.state || 'risk'} size={11} live />
          {tip && (
            <span
              style={{
                position: 'absolute', top: 24, left: -8, zIndex: 10,
                display: 'inline-flex', alignItems: 'center', gap: 7, whiteSpace: 'nowrap',
                padding: '7px 10px', border: '1px solid var(--border-mid)', borderRadius: 8,
                background: 'var(--surface-raised)', boxShadow: 'var(--shadow-tooltip)',
                fontFamily: 'var(--font-mono)', fontSize: 9.5, color: 'var(--text-body)',
              }}
            >
              <span style={{ fontWeight: 700, color: 'var(--risk)' }}>{status.label}</span>
              {status.detail}
            </span>
          )}
        </span>
      )}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4, minWidth: 0 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, height: 13 }}>
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: 13, fontWeight: 700, color: 'var(--ink)', whiteSpace: 'nowrap', lineHeight: '13px' }}>{name}</span>
          {rev && <Chip size="sm" uppercase={false}>{rev}</Chip>}
        </div>
        {addr && <div style={{ fontFamily: 'var(--font-mono)', fontSize: 9.5, color: 'var(--text-muted)', whiteSpace: 'nowrap', lineHeight: '9.5px' }}>{addr}</div>}
      </div>
      {right && <div style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: 12, flex: 'none' }}>{right}</div>}
    </header>
  );
}
