// Generic single-choice segmented control -- a mutually-exclusive input
// modeled as a native radio group (role="radiogroup" + role="radio"), not a
// tablist, per 26-UI-SPEC.md's Layout section resolution. Follows
// OptionsToggles.tsx's ToggleRow convention: plain hand-built markup with
// explicit ARIA attributes, no package.

export interface SegmentedOption<T extends string> {
  value: T;
  label: string;
}

export interface SegmentedControlProps<T extends string> {
  options: ReadonlyArray<SegmentedOption<T>>;
  value: T;
  onChange: (value: T) => void;
  ariaLabel: string;
}

function SegmentedControl<T extends string>({ options, value, onChange, ariaLabel }: SegmentedControlProps<T>) {
  const checkedIndex = options.findIndex((option) => option.value === value);

  return (
    <div
      className="ws-seg"
      role="radiogroup"
      aria-label={ariaLabel}
      onKeyDown={(event) => {
        if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return;
        event.preventDefault();

        const delta = event.key === 'ArrowLeft' ? -1 : 1;
        const nextIndex = checkedIndex + delta;
        if (nextIndex < 0 || nextIndex >= options.length) return; // no wrap past the ends

        const nextOption = options[nextIndex];
        onChange(nextOption.value);

        // Move DOM focus to the newly checked button in the same frame.
        const container = event.currentTarget;
        const buttons = container.querySelectorAll<HTMLButtonElement>('[role="radio"]');
        buttons[nextIndex]?.focus();
      }}
    >
      {options.map((option) => {
        const checked = option.value === value;
        return (
          <button
            key={option.value}
            type="button"
            role="radio"
            aria-checked={checked}
            tabIndex={checked ? 0 : -1}
            className={`ws-seg-option${checked ? ' ws-seg-option-active' : ''}`}
            onClick={() => onChange(option.value)}
          >
            {option.label}
          </button>
        );
      })}
    </div>
  );
}

export default SegmentedControl;
