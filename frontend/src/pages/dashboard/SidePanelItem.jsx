import { Link } from "@tanstack/react-router";

export default function SidePanelItem({ children, link, text, isPro }) {
  return (
    <div>
      <Link
        to={link}
        className="group flex items-center rounded-lg p-2 text-text_color hover:bg-hvr_gray"
      >
        <span
          className="flex-shrink-0 text-gray-500 transition duration-75 group-hover:text-text_color
            dark:text-gray-400"
        >
          {children}
        </span>
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
      </Link>
    </div>
  );
}
