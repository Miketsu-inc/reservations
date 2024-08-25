export default function SidePanelIcon({ children }) {
  return (
    <span
      className="flex-shrink-0 text-gray-500 transition duration-75 group-hover:text-gray-900
        dark:text-gray-400 dark:group-hover:text-white"
    >
      {children}
    </span>
  );
}
