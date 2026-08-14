import { useEffect, useRef } from 'react';

/**
 * The one implementation of the four behaviors every overlay in this app
 * needs: focus trap, Escape-to-close, scroll lock, and focus restore.
 *
 * Written for Phase 25's animated 260ms slide-over exit, not just this
 * phase's palette -- Phases 25, 26 and 27 import this hook rather than
 * reimplementing any of these four behaviors (24-CONTEXT.md). The single
 * decision that makes that work: the effect below is keyed on `[isOpen]`
 * alone, so its cleanup fires on the true->false transition while the
 * consumer is still mounted and still animating out, not only at unmount.
 * An empty-array effect would leave a still-animating consumer's page
 * scroll-locked for the whole exit.
 */

const FOCUSABLE_SELECTOR = 'a[href], button:not([disabled]), input, select, textarea, [tabindex]';

function getFocusableElements(container: HTMLElement): HTMLElement[] {
  return Array.from(container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)).filter(
    (el) => el.tabIndex !== -1
  );
}

export interface ModalBehaviorOptions {
  isOpen: boolean;
  onClose: () => void;
  initialFocusRef?: React.RefObject<HTMLElement | null>;
  scrollLockSelector?: string;
}

export interface ModalBehavior {
  containerRef: React.RefObject<HTMLDivElement>;
}

export function useModalBehavior({
  isOpen,
  onClose,
  initialFocusRef,
  scrollLockSelector,
}: ModalBehaviorOptions): ModalBehavior {
  const containerRef = useRef<HTMLDivElement>(null);

  // Consumers pass inline arrow functions for onClose. Reading it through a
  // ref (kept current by this tiny effect) means the main effect below can
  // omit onClose from its dependency array without going stale -- putting
  // it in the deps would re-run the whole effect, and re-lock the scroll /
  // re-capture the restore target, on every parent render.
  const onCloseRef = useRef(onClose);
  useEffect(() => {
    onCloseRef.current = onClose;
  }, [onClose]);

  useEffect(() => {
    if (!isOpen) return;

    const container = containerRef.current;

    // 1. Capture the restore target before focus moves, so it records
    // whatever the user was actually on -- the toolbar button, the rail
    // filter, wherever ⌘K fired -- rather than a hardcoded trigger.
    const restoreTarget =
      document.activeElement instanceof HTMLElement ? document.activeElement : null;

    // 2. Move initial focus synchronously, in this same frame -- no
    // timeout, so there is no frame where the overlay is visible but
    // unfocused.
    if (initialFocusRef?.current) {
      initialFocusRef.current.focus();
    } else if (container) {
      const [first] = getFocusableElements(container);
      (first ?? container).focus();
    }

    // 3. Lock scroll on the resolved element, saving its previous inline
    // overflow so cleanup restores it rather than blindly clearing it.
    const lockElement =
      document.querySelector<HTMLElement>(scrollLockSelector ?? '.ws-root') ?? document.body;
    const previousOverflow = lockElement.style.overflow;
    lockElement.style.overflow = 'hidden';

    // 4. One window listener covers both Escape and the focus trap. Escape
    // is checked first and fires regardless of which element inside the
    // overlay has focus.
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault();
        onCloseRef.current();
        return;
      }

      if (event.key === 'Tab' && container) {
        // Read fresh on every keypress -- the palette's row list changes
        // between keystrokes, so a list captured at open time would go stale.
        const focusable = getFocusableElements(container);
        if (focusable.length === 0) {
          event.preventDefault();
          container.focus();
          return;
        }

        const first = focusable[0];
        const last = focusable[focusable.length - 1];
        const active = document.activeElement;

        if (!active || !container.contains(active)) {
          event.preventDefault();
          first.focus();
        } else if (event.shiftKey && active === first) {
          event.preventDefault();
          last.focus();
        } else if (!event.shiftKey && active === last) {
          event.preventDefault();
          first.focus();
        }
      }
    };

    window.addEventListener('keydown', onKeyDown);

    return () => {
      window.removeEventListener('keydown', onKeyDown);
      lockElement.style.overflow = previousOverflow;
      if (restoreTarget && restoreTarget.isConnected) {
        restoreTarget.focus();
      }
    };
  }, [isOpen]);

  return { containerRef };
}
