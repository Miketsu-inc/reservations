import { Tick02Icon } from "@hugeicons/core-free-icons";
import Icon from "./Icon.jsx";

export default function CheckBox({ checked, styles, onChange, ...props }) {
  return (
    <div className="relative flex items-center justify-center">
      <input
        type="checkbox"
        checked={checked}
        onChange={onChange}
        className={`${styles} checked:border-primary checked:bg-primary
          bg-bg-input hover:not-checked:border-border-hover
          focus-visible:ring-primary-focus border-border-strong size-5
          cursor-pointer appearance-none rounded-md border outline-hidden
          transition-colors focus-visible:ring-2`}
        {...props}
      />
      {checked && (
        <Icon
          aria-hidden="true"
          icon={Tick02Icon}
          styles="pointer-events-none absolute size-4 text-fg-on-primary"
        />
      )}
    </div>
  );
}
