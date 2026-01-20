import { TickIcon } from "@reservations/assets";

export default function CheckBox({ checked, styles, onChange, ...props }) {
  return (
    <div className="relative flex items-center justify-center">
      <input
        type="checkbox"
        checked={checked}
        onChange={onChange}
        className={`${styles} checked:border-primary checked:bg-primary size-5
          cursor-pointer appearance-none rounded border-2 border-gray-400
          transition-colors dark:border-gray-500`}
        {...props}
      />
      {checked && <TickIcon styles="absolute size-5 text-white" />}
    </div>
  );
}
