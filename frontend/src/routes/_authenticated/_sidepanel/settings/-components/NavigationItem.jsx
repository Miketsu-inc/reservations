import { Link } from "@tanstack/react-router";

export default function NavigationItem({ path, label, children }) {
  return (
    <li className="group w-full">
      <Link
        to={path}
        activeProps={{
          className:
            "border-l-4 border-primary font-bold bg-hvr_gray/60 !text-text_color",
        }}
        className="flex items-center gap-2 rounded-md px-3 py-2 text-gray-600 hover:bg-hvr_gray/60
          dark:text-gray-400"
      >
        {children}
        {label}
      </Link>
    </li>
  );
}
