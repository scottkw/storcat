import type { ReactNode } from 'react';
import { useModalBehavior } from '../../hooks/useModalBehavior';

export interface DialogShellProps {
  isOpen: boolean;
  onClose: () => void;
  titleId: string;
  title: string;
  initialFocusRef?: React.RefObject<HTMLElement | null>;
  children: ReactNode;
  footer: ReactNode;
}

// The one shared 440px centred-dialog shell both this phase's dialogs use
// (rename here, delete-confirmation in 27-05) -- not two near-duplicate
// panel implementations (27-CONTEXT.md / 27-UI-SPEC.md).
//
// Always mounted by its consumer and returns null when closed -- the
// correct shape for a dialog (unlike Menu.tsx): the shared useModalBehavior
// hook observes the isOpen: true -> false transition to release scroll lock
// and restore focus, exactly as SettingsDialog documents.
function DialogShell({ isOpen, onClose, titleId, title, initialFocusRef, children, footer }: DialogShellProps) {
  const { containerRef } = useModalBehavior({ isOpen, onClose, initialFocusRef });

  if (!isOpen) return null;

  return (
    <div className="ws-dialog-scrim" onClick={onClose}>
      <div
        ref={containerRef}
        className="ws-dialog-panel"
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        onClick={(event) => event.stopPropagation()}
      >
        <div className="ws-dialog-header">
          <span id={titleId} className="ws-dialog-title">
            {title}
          </span>
          <button type="button" className="ws-dialog-close-x" aria-label="Close" onClick={onClose}>
            ×
          </button>
        </div>
        <div className="ws-dialog-body">{children}</div>
        <div className="ws-dialog-footer">{footer}</div>
      </div>
    </div>
  );
}

export default DialogShell;
