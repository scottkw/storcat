import { useEffect, useRef, useState } from 'react';
import { wailsAPI } from '../../../services/wailsAPI';
import { formatBytes } from '../../../lib/format';
import { sourceDisplayNameOf, type ScanSource } from '../../../types/scan';
import { volumes } from '../../../../wailsjs/go/models';

export interface VolumePickerProps {
  selected: ScanSource | null;
  onSelect: (source: ScanSource) => void;
}

// One card per detected volume, a chosen-folder pseudo-card once a folder is
// picked, and the always-present "choose any folder" link -- the only
// field-level gate on Create is having a source selected here at all
// (25-UI-SPEC.md's Source volume section). Owns its own volume enumeration
// (fetched fresh each time the form step mounts) so the staleness guard
// below lives beside the fetch it guards.
function VolumePicker({ selected, onSelect }: VolumePickerProps) {
  const [detected, setDetected] = useState<volumes.Volume[]>([]);

  // Mirrors CommandPalette's stale-response guard: a re-enumeration
  // triggered by a reopen must never let an older response land over a
  // newer one.
  const requestIdRef = useRef(0);

  // Kept in sync every render (not via an effect) so the one-time fetch
  // effect below can read the latest `selected` without re-subscribing --
  // preselection must never clobber a source the user (or a prior mount)
  // already chose.
  const selectedRef = useRef(selected);
  selectedRef.current = selected;

  useEffect(() => {
    const requestId = ++requestIdRef.current;
    wailsAPI.listVolumes().then((result) => {
      if (requestId !== requestIdRef.current) return; // superseded response, dropped
      if (!result.success) return;
      const list = result.volumes ?? [];
      setDetected(list);
      if (list.length > 0 && !selectedRef.current) {
        onSelect({ kind: 'volume', volume: list[0] });
      }
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function handleChooseFolder() {
    const result = await wailsAPI.selectDirectory();
    if (!result.success || !result.path) return;
    onSelect({ kind: 'folder', path: result.path });
  }

  return (
    <div className="ws-create-field">
      <label className="ws-create-label">Source volume</label>
      <div className="ws-create-vol-list">
        {detected.map((vol) => {
          const isSelected = selected?.kind === 'volume' && selected.volume.mountPath === vol.mountPath;
          return (
            <div
              key={vol.mountPath}
              className={`ws-create-vol-card${isSelected ? ' ws-create-vol-card-selected' : ''}`}
              role="button"
              tabIndex={0}
              onClick={() => onSelect({ kind: 'volume', volume: vol })}
              onKeyDown={(event) => {
                if (event.key === 'Enter' || event.key === ' ') {
                  event.preventDefault();
                  onSelect({ kind: 'volume', volume: vol });
                }
              }}
            >
              <span className="ws-create-vol-chip" aria-hidden="true" />
              <div className="ws-create-vol-text">
                <span className="ws-create-vol-name">{vol.name}</span>
                <span className="ws-create-vol-path mono">{vol.mountPath}</span>
              </div>
              <span className="ws-create-vol-size mono">{formatBytes(vol.totalBytes)}</span>
              <span className={`ws-create-tag${vol.readable ? '' : ' ws-create-tag-warn'}`}>
                {vol.readable ? 'mounted' : 'read errors'}
              </span>
            </div>
          );
        })}

        {selected?.kind === 'folder' && (
          <div
            className="ws-create-vol-card ws-create-vol-card-selected"
            role="button"
            tabIndex={0}
            onClick={handleChooseFolder}
            onKeyDown={(event) => {
              if (event.key === 'Enter' || event.key === ' ') {
                event.preventDefault();
                handleChooseFolder();
              }
            }}
          >
            <span className="ws-create-vol-chip ws-create-vol-chip-folder" aria-hidden="true">
              <svg
                width="11"
                height="11"
                viewBox="0 0 14 14"
                fill="none"
                stroke="var(--fn)"
                strokeWidth={1.4}
                aria-hidden="true"
                focusable="false"
              >
                <rect x="1" y="3" width="12" height="9" rx="1.5" />
                <line x1="1" y1="5.5" x2="13" y2="5.5" />
              </svg>
            </span>
            <div className="ws-create-vol-text">
              <span className="ws-create-vol-name">{sourceDisplayNameOf(selected)}</span>
              <span className="ws-create-vol-path mono">{selected.path}</span>
            </div>
          </div>
        )}
      </div>

      <button type="button" className="ws-create-folder-link" onClick={handleChooseFolder}>
        …or <span className="ws-create-folder-link-accent">choose any folder</span>
      </button>
    </div>
  );
}

export default VolumePicker;
