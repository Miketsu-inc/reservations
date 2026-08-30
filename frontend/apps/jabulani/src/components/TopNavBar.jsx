import { Link } from "@tanstack/react-router";

export function TopNavBar({ children }) {
  return (
    <nav className="border-border_color h-fit w-full border-b">
      <div className="overflow-x-auto overflow-y-hidden">
        <ul className="flex flex-row items-center gap-2 px-4">{children}</ul>
      </div>
    </nav>
  );
}

export function TopNavBarItem({ from, to, children }) {
  return (
    <li className="flex flex-col gap-2 rounded-lg">
      <Link
        className="hover:bg-hvr_gray after:border-primary relative mb-2 flex
          flex-row items-center gap-2 rounded-lg px-2 py-1.5 after:absolute
          after:-bottom-2 after:left-0 after:w-full after:border-b-2 after:pt-4
          after:opacity-0"
        from={from}
        to={to}
        activeProps={{
          className: "after:opacity-100",
        }}
      >
        {children}
      </Link>
    </li>
  );
}
