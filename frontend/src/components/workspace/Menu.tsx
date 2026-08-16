import { Fragment, useEffect, useLayoutEffect, useRef, useState } from 'react';
import { useModalBehavior } from '../../hooks/useModalBehavior';

export interface MenuItemSpec {
  id: string;
  label: string;
  danger?: boolean;
  dividerBefore?: boolean;
  onSelect: () => void;
}

export interface MenuProps {
  isOpen: boolean;
  onClose: () => void;
  triggerRef: React.RefObject<HTMLButtonElement | null>;
  items: MenuItemSpec[];
  id: string;
  ariaLabel: string;
}

interface MenuPosition {
  top: number;
  right: number;
}

// The app's first menu primitive. Conditionally mounted by its parent (the
// parent renders this component only while open), so it may assume isOpen
// is true whenever it renders and must NOT carry an internal early-return
// mount gate keyed on isOpen -- that is SettingsDialog's always-mounted
// shape, the wrong one to copy here (27-PATTERNS.md). A menu has no exit
// animation whose teardown the shared hook would need to observe.
function Menu({ isOpen, onClose, triggerRef, items, id, ariaLabel }: MenuProps) {
  const itemRefs = useRef<Array<HTMLButtonElement | null>>([]);
  const firstItemRef = useRef<HTMLButtonElement | null>(null);
  const [activeIndex, setActiveIndex] = useState(0);
  const [position, setPosition] = useState<MenuPosition | null>(null);

  // scrollLockSelector deliberately points at the trigger button itself
  // (.ws-details-overflow), not .ws-root -- the hook always sets
  // overflow: hidden on *some* element, so aiming it at a 22x22 icon button
  // with no internal scroll makes that assignment a functional no-op. This
  // is configuration, not a fork of the hook (27-CONTEXT.md).
  const { containerRef } = useModalBehavior({
    isOpen,
    onClose,
    initialFocusRef: firstItemRef,
    scrollLockSelector: '.ws-details-overflow',
  });

  // Measured once from the trigger's own screen position, in the same
  // frame, before paint -- correct whether the details panel sits on the
  // app's left or right side under the rail-position setting. No
  // viewport-collision or reopen-upward-when-crowded logic is implemented:
  // three short items opening downward from a trigger near the top of a
  // bounded-height panel has no realistic path to overflowing the bottom of
  // the window at any size this app supports (backstop must_have, reasoned
  // from the app's real layout geometry, not proven by an automated
  // boundary sweep).
  useLayoutEffect(() => {
    if (!isOpen) return;
    const trigger = triggerRef.current;
    if (!trigger) return;
    const rect = trigger.getBoundingClientRect();
    setPosition({ top: rect.bottom + 6, right: window.innerWidth - rect.right });
  }, [isOpen]);

  // Click-outside is not provided by useModalBehavior (no scrim, no pointer
  // listener) -- add it here. The trigger is excluded so its own onClick
  // toggle-closed doesn't race this listener into a close-then-reopen.
  useEffect(() => {
    if (!isOpen) return undefined;
    const handlePointerDown = (event: PointerEvent) => {
      const target = event.target as Node;
      if (containerRef.current?.contains(target)) return;
      if (triggerRef.current?.contains(target)) return;
      // CR-01: suppress the browser's own compatibility mousedown (and its
      // focus-follows-click default action) for this interaction, so
      // useModalBehavior's restoreTarget.focus() is the only focus mutation
      // that happens. Without this, mousedown fires after this handler and
      // blurs focus to <body>, overwriting the restore that already ran.
      event.preventDefault();
      onClose();
    };
    document.addEventListener('pointerdown', handlePointerDown);
    return () => document.removeEventListener('pointerdown', handlePointerDown);
  }, [isOpen, onClose]);

  return (
    <div
      ref={containerRef}
      className="ws-menu"
      role="menu"
      id={id}
      aria-label={ariaLabel}
      style={{ top: position?.top, right: position?.right }}
      onKeyDown={(event) => {
        if (event.key === 'ArrowDown') {
          event.preventDefault();
          const next = (activeIndex + 1) % items.length;
          setActiveIndex(next);
          itemRefs.current[next]?.focus();
        } else if (event.key === 'ArrowUp') {
          event.preventDefault();
          const next = (activeIndex - 1 + items.length) % items.length;
          setActiveIndex(next);
          itemRefs.current[next]?.focus();
        } else if (event.key === 'Enter' || event.key === ' ') {
          event.preventDefault();
          items[activeIndex]?.onSelect();
        }
      }}
    >
      {items.map((item, index) => (
        <Fragment key={item.id}>
          {item.dividerBefore && <hr className="ws-menu-divider" role="none" />}
          <button
            type="button"
            role="menuitem"
            ref={(el) => {
              itemRefs.current[index] = el;
              if (index === 0) firstItemRef.current = el;
            }}
            tabIndex={index === activeIndex ? 0 : -1}
            className={item.danger ? 'ws-menu-item ws-menu-item-danger' : 'ws-menu-item'}
            onClick={() => item.onSelect()}
          >
            {item.label}
          </button>
        </Fragment>
      ))}
    </div>
  );
}

export default Menu;
