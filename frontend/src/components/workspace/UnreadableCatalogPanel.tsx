import { useAppContext } from '../../contexts/AppContext';

// A catalog can arrive broken by either of two independent routes: the rail
// listing's own parse-status check (BrowseCatalogs' detectParseError, which
// formats a JSON syntax error as "byte N: reason" and carries a real byte
// offset), or the tree load itself failing (LoadCatalogFlat -- covers a file
// deleted between listing and click, a permission error, or valid JSON that
// isn't a catalog, none of which carry a byte offset). The listing error is
// preferred when both exist because it is the richer diagnostic.
function pickRawError(catalog: { parseError: string } | undefined, treeMessage: string): string {
  const listingError = catalog?.parseError ?? '';
  return listingError !== '' ? listingError : treeMessage;
}

const BYTE_OFFSET_RE = /^byte (\d+): (.*)$/s;
const READ_FAILURE_RE = /no such file or directory|permission denied|failed to read catalog file/i;

function UnreadableCatalogPanel() {
  const { state } = useAppContext();
  const catalog = state.catalogs.find((c) => c.path === state.currentCatalogId);
  const treeMessage = state.tree.status === 'error' ? state.tree.message : '';
  const rawError = pickRawError(catalog, treeMessage);

  if (rawError === '') return null;

  const byteMatch = BYTE_OFFSET_RE.exec(rawError);
  // A byte offset means the string is already the "byte N: reason" shape
  // detectParseError produces -- the reason is everything after that prefix.
  // No offset means the failure never got that far: a missing file, a
  // permission error, or (rarely) a structural check after a successful
  // parse -- "Failed at" is an em dash in all of those cases, per FPA-23-05-A.
  const failedAt = byteMatch ? `byte ${byteMatch[1]}` : '—';
  const reason = byteMatch ? byteMatch[2] : rawError.replace(/^load catalog for flatten: /, '');
  const isReadFailure = !byteMatch && READ_FAILURE_RE.test(rawError);
  const parser = isReadFailure ? 'not reached' : 'v2 object / v1 array';

  const metaRows: [string, string][] = [
    ['File', catalog?.filename ?? '—'],
    ['Failed at', failedAt],
    ['Reason', reason],
    ['Parser', parser],
  ];

  return (
    <div
      className="pane-scroll"
      style={{ padding: '24px 22px', display: 'flex', flexDirection: 'column', gap: 18 }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 11 }}>
        <span
          aria-hidden="true"
          style={{
            width: 22,
            height: 22,
            borderRadius: 6,
            background: 'rgba(229,83,75,.16)',
            color: '#e5534b',
            fontSize: 13,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flex: 'none',
          }}
        >
          !
        </span>
        <span style={{ fontSize: 15, fontWeight: 600 }}>This catalog can&rsquo;t be read</span>
      </div>

      <span style={{ fontSize: 12.5, lineHeight: 1.6, color: 'var(--dm)', maxWidth: 520 }}>
        This catalog&rsquo;s JSON couldn&rsquo;t be parsed. The .html copy may still open, and the volume
        can be re-scanned later to rebuild it.
      </span>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 1, maxWidth: 520 }}>
        {metaRows.map(([label, value]) => (
          <div
            key={label}
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              gap: 12,
              padding: '7px 0',
              borderBottom: '1px solid var(--l2)',
              fontSize: 11.5,
            }}
          >
            <span style={{ color: 'var(--dm)' }}>{label}</span>
            <span className="mono">{value}</span>
          </div>
        ))}
      </div>

      {/* Verbatim, unedited -- never truncated, paraphrased or summarised.
          No max height, no internal scroll: a single-line parser message is
          the expected shape (23-UI-SPEC E4 overflow). */}
      <div
        className="mono"
        style={{
          padding: 12,
          borderRadius: 9,
          background: 'var(--ch)',
          fontSize: 11.5,
          color: '#e5534b',
          wordBreak: 'break-all',
        }}
      >
        {rawError}
      </div>

      {/* No action buttons this phase -- STATE-03's re-scan / open-HTML /
          remove-from-library trio lands together in Phase 28. */}
    </div>
  );
}

export default UnreadableCatalogPanel;
