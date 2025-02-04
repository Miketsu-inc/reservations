import { Link } from "@tanstack/react-router";

export default function NavigationItem({ path, label, children }) {
  return (
    <li className="w-full">
      <Link
        to={path}
        activeProps={{
          style: {
            borderLeft: "4px solid #2563eb", // Example of applying border style
            backgroundColor: "rgba(31, 41, 55, 0.5)", // Background color when active
            fontWeight: "550", // Example of changing font weight
          },
        }}
        className="flex items-center gap-2 rounded-md px-3 py-2 text-text_color
          hover:bg-hvr_gray/60"
      >
        {children}
        {label}
      </Link>
    </li>
  );
}
