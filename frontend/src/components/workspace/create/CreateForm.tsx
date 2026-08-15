import { slugifyRoot, willWritePaths } from '../../../lib/scanFormat';
import { sourceDisplayNameOf, type ScanSource } from '../../../types/scan';

export interface CreateFormProps {
  source: ScanSource | null;
  title: string;
  onTitleChange: (value: string) => void;
  root: string;
  onRootChange: (value: string) => void;
  catalogDir: string;
  writeHTML: boolean;
  secondaryDir: string;
  // Filenames (e.g. "root.json") already on disk in catalogDir, per the
  // rail's already-loaded listing -- the cheapest correct existence check,
  // per this task's explicit "no new binding" instruction.
  existingCatalogFilenames: Set<string>;
}

// Title and filename root are independent fields: each shows the source's
// derived default purely as a native placeholder (visible only while the
// field is empty), so typing in one can never overwrite the other and a
// cleared field always falls back to the *current* source's default rather
// than a stale one (CRT-04 adjacency). Both fields, and the WILL WRITE
// preview below them, recompute synchronously from `source`/toggle props on
// every render -- there is no async gap and no manual refresh.
function CreateForm({
  source,
  title,
  onTitleChange,
  root,
  onRootChange,
  catalogDir,
  writeHTML,
  secondaryDir,
  existingCatalogFilenames,
}: CreateFormProps) {
  const displayName = source ? sourceDisplayNameOf(source) : '';
  const titlePlaceholder = displayName || 'Untitled catalog';
  const rootPlaceholder = (displayName && slugifyRoot(displayName)) || 'catalog';
  const effectiveRoot = root.trim() || rootPlaceholder;

  const willWrite = willWritePaths({
    catalogDir,
    root: effectiveRoot,
    writeHtml: writeHTML,
    secondaryDir: secondaryDir || undefined,
  });

  return (
    <>
      <div className="ws-create-grid">
        <div className="ws-create-field">
          <label className="ws-create-label">Catalog title</label>
          <input
            className="ws-create-input"
            value={title}
            onChange={(event) => onTitleChange(event.target.value)}
            placeholder={titlePlaceholder}
          />
        </div>
        <div className="ws-create-field">
          <label className="ws-create-label">Filename root</label>
          <div className="ws-create-input-suffix-row">
            <input
              className="ws-create-input mono"
              value={root}
              onChange={(event) => onRootChange(event.target.value)}
              placeholder={rootPlaceholder}
            />
            <span className="ws-create-suffix mono">.json / .html</span>
          </div>
        </div>
      </div>

      <div className="ws-create-field">
        <label className="ws-create-label">Will write</label>
        <div className="ws-create-willwrite mono">
          {willWrite.map((path) => {
            const filename = path.slice(path.lastIndexOf('/') + 1);
            const inCatalogDir = path.startsWith(`${catalogDir.replace(/[\\/]+$/, '')}/`);
            const alreadyExists = inCatalogDir && existingCatalogFilenames.has(filename);
            return (
              <div className="ws-create-willwrite-row" key={path}>
                {path}
                {alreadyExists && <span className="ws-create-willwrite-exists"> — already exists</span>}
              </div>
            );
          })}
        </div>
      </div>
    </>
  );
}

export default CreateForm;
