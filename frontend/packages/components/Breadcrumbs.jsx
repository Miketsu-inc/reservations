import { ArrowLeft01Icon } from "@hugeicons/core-free-icons";
import { isMatch, Link, useMatches } from "@tanstack/react-router";
import Icon from "./Icon.jsx";

export default function Breadcrumbs() {
  const matches = useMatches();
  const lastMatch = matches[matches.length - 1];

  const matchesWithCrumbs = matches.filter((match) =>
    isMatch(match, "loaderData.crumb")
  );
  const items = matchesWithCrumbs.map(({ pathname, loaderData }) => {
    return {
      href: pathname,
      label: loaderData?.crumb,
    };
  });

  return (
    <nav aria-label="breadcrumb">
      <ol className="wrap-break-words flex flex-wrap items-center gap-2">
        {items.map((item, index) => (
          <li
            className="hover:text-text_color inline-flex items-center gap-2"
            key={item.href}
          >
            {lastMatch?.pathname === item.href ? (
              <span aria-current="page">{item.label}</span>
            ) : (
              <Link to={item.href}>{item.label}</Link>
            )}
            {index < items.length - 1 && (
              <Icon icon={ArrowLeft01Icon} styles="rotate-180 size-4" />
            )}
          </li>
        ))}
      </ol>
    </nav>
  );
}
