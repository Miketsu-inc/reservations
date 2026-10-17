import { Link } from "@tanstack/react-router";

export function TopNavBar({ children }) {
  return (
    <nav className="border-border-subtle h-fit w-full min-w-0 shrink-0 border-b">
      <div
        className="w-full min-w-0 scrollbar-thin overflow-x-auto
          overflow-y-hidden"
      >
        <ul
          className="flex w-max min-w-full flex-row flex-nowrap items-center
            gap-1 px-3"
        >
          {children}
        </ul>
      </div>
    </nav>
  );
}

export function TopNavBarItem({ styles, from, to, children, ...props }) {
  return (
    <li className="flex shrink-0 flex-col gap-2 rounded-lg">
      <Link
        className={`${styles} text-fg-muted hover:bg-hover hover:text-fg
          after:border-primary relative mb-2 flex flex-row items-center gap-2
          rounded-lg px-2.5 py-1.5 text-sm font-medium whitespace-nowrap
          after:absolute after:-bottom-2 after:left-0 after:w-full
          after:border-b-2 after:pt-4 after:opacity-0`}
        from={from}
        to={to}
        activeProps={
          props?.activeProps ?? {
            className: "text-fg! after:opacity-100",
          }
        }
        {...props}
      >
        {children}
      </Link>
    </li>
  );
}
