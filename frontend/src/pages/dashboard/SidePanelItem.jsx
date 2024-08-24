export default function SidePanelItem({ children, link, text, isPro }) {
  return (
    <li>
      <a
        href={link}
        className="group flex items-center rounded-lg p-2 text-gray-900 hover:bg-gray-100
          dark:text-white dark:hover:bg-gray-700"
      >
        {children}
        <span className="ms-3 flex-1 whitespace-nowrap">{text}</span>
        {isPro ? (
          <span
            className="ms-3 inline-flex items-center justify-center rounded-full bg-gray-100 px-2
              text-sm font-medium text-gray-800 dark:bg-gray-700 dark:text-gray-300"
          >
            Pro
          </span>
        ) : (
          <></>
        )}
      </a>
    </li>
  );
}
