export default function SelectorItem({
  value,
  children,
  onClick,
  styles,
  key,
}) {
  return (
    <li>
      <button
        key={key}
        type="button"
        className={`inline-flex w-full px-4 py-2 text-sm text-gray-600 hover:bg-hvr_gray
          hover:text-text_color dark:text-gray-300 dark:hover:text-text_color ${styles} `}
        role="menuitem"
        onClick={() => onClick(value)}
      >
        <span className="inline-flex items-center">{children}</span>
      </button>
    </li>
  );
}
