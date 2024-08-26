export default function SidePanelItem({ children, link, text, isPro }) {
  return (
    <div>
      <a
        href={link}
        className="hover:bg-hvr-gray text-text-color group flex items-center rounded-lg p-2"
      >
        {children}
        <span className="ms-3 flex-1 whitespace-nowrap">{text}</span>
        {isPro ? (
          <span
            className="ms-3 inline-flex items-center justify-center rounded-full bg-gray-300 px-2
              text-sm font-medium text-gray-800 dark:bg-gray-700 dark:text-gray-300"
          >
            Pro
          </span>
        ) : (
          <></>
        )}
      </a>
    </div>
  );
}
