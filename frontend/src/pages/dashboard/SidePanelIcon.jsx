export default function SidePanelIcon({ children }) {
  return (
    <span
      className="group-hover:text-text-color flex-shrink-0 text-gray-500 transition duration-75
        dark:text-gray-400"
    >
      {children}
    </span>
  );
}
