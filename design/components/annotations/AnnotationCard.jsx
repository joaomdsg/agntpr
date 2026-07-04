import React from 'react';
import { Chip } from '../chips/Chip.jsx';

const AUTHOR_COLORS = { you: 'var(--signal)', agent: 'var(--agent)' };
const SEV_STATE = { blocking: 'risk', 'self-flag': 'agent', question: 'held', nit: 'neutral' };

/**
 * The annotation card — authorship is the 3px left border (you = cyan,
 * agent = purple). This left-accent pattern is native to annotations;
 * don't generalize it to other card types.
 */
export function AnnotationCard({ author = 'you', sev = 'nit', where, scope, actions, onClick, style, children }) {
  const col = AUTHOR_COLORS[author] || 'var(--signal)';
  return (
    <div
      onClick={onClick}
      style={{
        border: '1px solid var(--hairline)',
        borderLeft: '3px solid ' + col,
        borderRadius: 'var(--r-ann)',
        background: 'var(--surface-raised)',
        padding: '10px 11px',
        cursor: onClick ? 'pointer' : 'default',
        ...style,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 7, marginBottom: 6 }}>
        <Chip state={author === 'agent' ? 'agent' : 'you'} dot>{author}</Chip>
        <Chip state={SEV_STATE[sev] || 'neutral'}>{sev}</Chip>
        {where && <span style={{ marginLeft: 'auto', fontFamily: 'var(--font-mono)', fontSize: 9.5, color: 'var(--text-faint)' }}>{where}</span>}
      </div>
      <div style={{ fontFamily: 'var(--font-ui)', fontSize: 11.5, lineHeight: 1.5, color: 'var(--text-body)' }}>{children}</div>
      {scope && <div style={{ fontFamily: 'var(--font-mono)', fontSize: 9.5, color: 'var(--text-faint)', marginTop: 7 }}>{scope}</div>}
      {actions && <div style={{ fontFamily: 'var(--font-mono)', fontSize: 9.5, color: 'var(--signal)', marginTop: 7 }}>{actions}</div>}
    </div>
  );
}
