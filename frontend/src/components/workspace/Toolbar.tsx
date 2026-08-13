import React, { useLayoutEffect } from 'react';
import storcatIcon from '../../storcat-icon.svg';
import { Environment } from '../../../wailsjs/runtime/runtime';

export interface ToolbarProps {
  themeName: string;
  showDetailsChip: boolean;
  detailsOpen: boolean;
  onToggleDetails: () => void;
}

function Toolbar({ themeName, showDetailsChip, detailsOpen, onToggleDetails }: ToolbarProps) {
  // useLayoutEffect (not useEffect) so this fires before the browser paints,
  // closing React's own commit-to-paint gap. Environment() itself is still
  // async, so this narrows but does not eliminate the one-frame window.
  useLayoutEffect(() => {
    let cancelled = false;
    // Environment() dereferences window.runtime synchronously, so outside a Wails
    // webview it THROWS rather than rejecting — which would take down the whole tree.
    // Deferring through a resolved promise turns that into a rejection the catch below
    // already handles, so the shell still renders in a plain browser (needed for UAT).
    Promise.resolve()
      .then(() => Environment())
      .then((env) => {
        if (cancelled) return;
        if (env.platform === 'darwin') {
          document.documentElement.style.setProperty('--toolbar-inset-left', '78px');
        }
      })
      .catch(() => {
        // Platform query failed -- leave the toolbar inset at its 0px default.
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div
      className="ws-toolbar"
      style={
        {
          '--wails-draggable': 'drag',
        } as React.CSSProperties & { '--wails-draggable'?: string }
      }
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexShrink: 0 }}>
        <img
          src={storcatIcon}
          alt=""
          aria-hidden="true"
          style={{ width: 16, height: 16, borderRadius: 4 }}
        />
        <span style={{ fontSize: 13, fontWeight: 600, letterSpacing: '-0.01em' }}>StorCat</span>
      </div>

      <div style={{ flex: 1, minWidth: 0, display: 'flex', justifyContent: 'center' }}>
        <button
          type="button"
          className="no-drag ws-search"
          aria-label="Search every catalog"
          style={{
            width: '100%',
            maxWidth: 460,
            minWidth: 0,
            height: 30,
            borderRadius: 8,
            background: 'var(--bg)',
            border: '1px solid var(--l)',
            display: 'flex',
            alignItems: 'center',
            gap: 9,
            padding: '0 10px',
            cursor: 'pointer',
            color: 'inherit',
            font: 'inherit',
          }}
        >
          <svg
            width="13"
            height="13"
            viewBox="0 0 16 16"
            fill="none"
            stroke="var(--dm)"
            strokeWidth={1.6}
            aria-hidden="true"
            focusable="false"
          >
            <circle cx="7" cy="7" r="4.5" />
            <line x1="10.5" y1="10.5" x2="14" y2="14" />
          </svg>
          <span style={{ fontSize: 12.5, color: 'var(--dm)' }}>Search every catalog…</span>
          <span
            className="mono"
            style={{
              marginLeft: 'auto',
              fontSize: 11,
              color: 'var(--fn)',
              border: '1px solid var(--l)',
              borderRadius: 4,
              padding: '1px 5px',
            }}
          >
            ⌘K
          </span>
        </button>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexShrink: 0 }}>
        {showDetailsChip && (
          <button
            type="button"
            className="no-drag"
            aria-expanded={detailsOpen}
            onClick={onToggleDetails}
            style={{
              fontSize: 11.5,
              border: '1px solid var(--l)',
              borderRadius: 6,
              padding: '3px 8px',
              background: 'transparent',
              cursor: 'pointer',
              whiteSpace: 'nowrap',
              color: detailsOpen ? 'var(--ac)' : 'var(--dm)',
            }}
          >
            Details
          </button>
        )}

        <button
          type="button"
          className="no-drag ws-chip"
          style={{
            fontSize: 11.5,
            color: 'var(--dm)',
            border: '1px solid var(--l)',
            borderRadius: 6,
            padding: '3px 8px',
            background: 'transparent',
            cursor: 'pointer',
            whiteSpace: 'nowrap',
          }}
        >
          {themeName}
        </button>

        <button
          type="button"
          className="no-drag"
          aria-label="Settings"
          style={{
            background: 'transparent',
            border: 'none',
            padding: 0,
            display: 'flex',
            cursor: 'pointer',
          }}
        >
          <svg
            width="15"
            height="15"
            viewBox="0 0 16 16"
            fill="none"
            stroke="var(--dm)"
            strokeWidth={1.4}
            aria-hidden="true"
            focusable="false"
          >
            <circle cx="8" cy="8" r="2.4" />
            <circle cx="8" cy="8" r="6" />
          </svg>
        </button>
      </div>
    </div>
  );
}

export default Toolbar;
