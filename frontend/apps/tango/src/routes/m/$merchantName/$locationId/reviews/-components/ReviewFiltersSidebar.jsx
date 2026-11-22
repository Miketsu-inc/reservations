import { Avatar, CheckBox, SearchInput } from "@reservations/components";
import { formatShortName, toAvatar } from "@reservations/lib";
import { useState } from "react";
import RatingSummary from "../../../-components/RatingSummary";

const COLLAPSED_COUNT = 5;

function FilterGroup({
  title,
  items,
  selectedIds,
  onChange,
  searchPlaceholder,
  emptyText,
  renderLabel,
  renderMeta,
}) {
  const [search, setSearch] = useState("");
  const [showAll, setShowAll] = useState(false);

  const term = search.trim().toLowerCase();
  const filteredItems = term
    ? items.filter((item) => item.name.toLowerCase().includes(term))
    : items;
  const visibleItems = showAll
    ? filteredItems
    : filteredItems.slice(0, COLLAPSED_COUNT);
  const hiddenCount = filteredItems.length - COLLAPSED_COUNT;

  function toggle(id) {
    onChange(
      selectedIds.includes(id)
        ? selectedIds.filter((selectedId) => selectedId !== id)
        : [...selectedIds, id]
    );
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <h3 className="text-base font-semibold">{title}</h3>
        {selectedIds.length > 0 && (
          <button
            type="button"
            onClick={() => onChange([])}
            className="text-fg-muted hover:text-fg cursor-pointer text-sm"
          >
            Clear
          </button>
        )}
      </div>

      <SearchInput
        placeholder={searchPlaceholder}
        searchText={search}
        onChange={setSearch}
      />

      <div className="flex flex-col gap-2">
        {visibleItems.length > 0 ? (
          visibleItems.map((item) => (
            <label
              key={item.id}
              className="hover:bg-hover group flex cursor-pointer items-center
                justify-between gap-2 rounded-lg px-2 py-1.5 transition-colors"
            >
              <div className="flex min-w-0 items-center gap-3">
                <CheckBox
                  checked={selectedIds.includes(item.id)}
                  onChange={() => toggle(item.id)}
                />
                {renderLabel?.(item)}
                <span
                  className="text-fg-secondary group-hover:text-fg truncate
                    text-sm"
                >
                  {item.name}
                </span>
              </div>
              {renderMeta(item)}
            </label>
          ))
        ) : (
          <p className="text-fg-muted py-1 text-sm">{emptyText}</p>
        )}

        {hiddenCount > 0 && (
          <button
            type="button"
            onClick={() => setShowAll((prev) => !prev)}
            className="text-fg-muted hover:text-fg w-fit cursor-pointer text-sm"
          >
            {showAll ? "Show less" : `Show ${hiddenCount} more`}
          </button>
        )}
      </div>
    </div>
  );
}

export default function ReviewFiltersSidebar({
  services,
  serviceCounts,
  team,
  filters,
  onChange,
}) {
  const isSticky = services.length + team.length <= 9;

  const teamItems = team.map((member) => ({
    ...member,
    ...toAvatar(member),
    name: formatShortName(member),
  }));

  return (
    <div className={`flex flex-col gap-8 ${isSticky ? "sticky top-20" : ""}`}>
      <FilterGroup
        title="Service"
        items={services}
        selectedIds={filters.serviceIds}
        onChange={(serviceIds) => onChange({ serviceIds })}
        searchPlaceholder="Search services"
        emptyText="No services found"
        renderMeta={(service) => (
          <span className="text-fg-muted shrink-0 text-sm">
            {serviceCounts[service.id] ?? 0}
          </span>
        )}
      />

      <FilterGroup
        title="Team"
        items={teamItems}
        selectedIds={filters.employeeIds}
        onChange={(employeeIds) => onChange({ employeeIds })}
        searchPlaceholder="Search team"
        emptyText="No team members found"
        renderLabel={(member) => (
          <Avatar
            styles="size-6! text-[10px]! shrink-0"
            img={member.img}
            initials={member.initials}
          />
        )}
        renderMeta={(member) =>
          member.review_count > 0 && (
            <RatingSummary
              average={member.average_rating}
              styles="shrink-0 text-xs"
            >
              <span className="text-fg-muted">({member.review_count})</span>
            </RatingSummary>
          )
        }
      />
    </div>
  );
}
