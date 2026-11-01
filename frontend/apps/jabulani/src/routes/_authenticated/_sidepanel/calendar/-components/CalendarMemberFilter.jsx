import { ArrowDown01Icon, User03Icon } from "@hugeicons/core-free-icons";
import {
  Avatar,
  AvatarStack,
  CheckBox,
  Icon,
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@reservations/components";
import { toAvatar } from "@reservations/lib";
import { useRef, useState } from "react";

// Stands in for bookings and blocked times without anyone on the team.
export const UNASSIGNED_ID = "unassigned";

export function UnassignedAvatar({ styles }) {
  return (
    <div
      className={`bg-bg-inactive flex shrink-0 items-center justify-center
        rounded-full ${styles}`}
    >
      <Icon icon={User03Icon} styles="size-1/2 text-fg-muted" />
    </div>
  );
}

function getTriggerText(
  selectedMembers,
  selected,
  isEveryone,
  currentEmployeeId
) {
  if (isEveryone) return "Everyone";
  if (selected.size > 1) return `${selected.size} selected`;
  if (selectedMembers.length === 0) return "Unassigned";

  const [member] = selectedMembers;
  return member.id === currentEmployeeId ? "Only me" : toAvatar(member).name;
}

function FilterRow({ isSelected, onToggle, onOnly, children }) {
  return (
    <li className="group relative flex items-center">
      <button
        type="button"
        data-filter-item
        aria-pressed={isSelected}
        onClick={onToggle}
        className={`hover:bg-hover focus-visible:bg-hover flex min-w-0 flex-1
          cursor-pointer items-center gap-3 rounded-md px-2 py-1.5 text-left
          text-sm outline-hidden ${onOnly ? "pointer-fine:pr-14" : ""}`}
      >
        <CheckBox
          checked={isSelected}
          readOnly
          tabIndex={-1}
          styles="pointer-events-none"
        />
        {children}
      </button>
      {/* Shown on hover, and always on touch where there is no hover. */}
      {onOnly && (
        <button
          type="button"
          onClick={onOnly}
          className="text-fg-muted hover:text-fg focus-visible:bg-hover absolute
            right-2 cursor-pointer rounded-md px-2 py-1 text-xs font-medium
            opacity-0 outline-hidden group-hover:opacity-100
            focus-visible:opacity-100 pointer-coarse:static
            pointer-coarse:opacity-100"
        >
          Only
        </button>
      )}
    </li>
  );
}

/**
 * Picks whose bookings the calendar shows. At least one option always stays
 * selected, since an empty calendar has nothing to show or create on.
 */
export default function CalendarMemberFilter({
  members,
  selectedIds,
  showUnassigned,
  onChange,
  currentEmployeeId,
}) {
  const [isOpen, setIsOpen] = useState(false);
  const listRef = useRef(null);

  const optionIds = members.map((member) => member.id);
  if (showUnassigned) optionIds.push(UNASSIGNED_ID);

  const selected = new Set(selectedIds);
  const selectedMembers = members.filter((member) => selected.has(member.id));
  const isEveryone = optionIds.every((id) => selected.has(id));

  function toggle(id) {
    if (!selected.has(id)) {
      onChange([...selectedIds, id]);
      return;
    }

    if (selectedIds.length > 1) {
      onChange(selectedIds.filter((selectedId) => selectedId !== id));
    }
  }

  // Lets the arrow keys move between the toggles while Tab still reaches the
  // "Only" buttons.
  function handleKeyDown(e) {
    if (e.key !== "ArrowDown" && e.key !== "ArrowUp") return;
    e.preventDefault();

    const items = [...listRef.current.querySelectorAll("[data-filter-item]")];
    const index = items.indexOf(document.activeElement);
    const next =
      e.key === "ArrowDown"
        ? Math.min(index + 1, items.length - 1)
        : Math.max(index - 1, 0);
    items[next]?.focus();
  }

  if (members.length === 0) return null;

  const triggerText = getTriggerText(
    selectedMembers,
    selected,
    isEveryone,
    currentEmployeeId
  );

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          aria-label={`Shown team members: ${triggerText}`}
          title={triggerText}
          className="bg-bg-input border-border-strong
            hover:not-data-popup-open:border-border-hover
            focus-visible:border-primary focus-visible:ring-primary-focus
            data-popup-open:border-primary data-popup-open:ring-primary-focus
            flex shrink-0 cursor-pointer items-center gap-1 rounded-lg border
            py-1.5 pr-1.5 pl-2 outline-hidden
            transition-[border-color,box-shadow] duration-150
            focus-visible:ring-3 data-popup-open:ring-3"
        >
          {selectedMembers.length > 0 ? (
            <AvatarStack
              avatars={selectedMembers.map((member) => toAvatar(member))}
            />
          ) : (
            <UnassignedAvatar styles="size-6" />
          )}
          <Icon
            icon={ArrowDown01Icon}
            styles={`text-fg-muted size-4 transition-transform
              ${isOpen ? "rotate-180" : ""}`}
          />
        </button>
      </PopoverTrigger>
      <PopoverContent styles="w-72! p-1!" align="end">
        <p className="text-fg-muted px-2 pt-1 pb-2 text-xs font-medium">
          Show bookings for
        </p>
        <ul
          ref={listRef}
          onKeyDown={handleKeyDown}
          className="max-h-80 overflow-y-auto dark:scheme-dark"
        >
          <FilterRow
            isSelected={isEveryone}
            onToggle={() => onChange(optionIds)}
          >
            <span className="font-medium">Everyone</span>
          </FilterRow>
          {members.map((member) => {
            const isSelected = selected.has(member.id);
            const { img, initials, name } = toAvatar(member);

            return (
              <FilterRow
                key={member.id}
                isSelected={isSelected}
                onToggle={() => toggle(member.id)}
                onOnly={() => onChange([member.id])}
              >
                <Avatar
                  img={img}
                  initials={initials}
                  alt={name}
                  styles={`size-6! text-[10px]! shrink-0
                  ${isSelected ? "" : "opacity-40 grayscale"}`}
                />
                <span
                  className={`min-w-0 truncate
                  ${isSelected ? "text-fg" : "text-fg-muted"}`}
                >
                  {name}
                  {member.id === currentEmployeeId && (
                    <span className="text-fg-muted"> (You)</span>
                  )}
                </span>
              </FilterRow>
            );
          })}
          {showUnassigned && (
            <FilterRow
              isSelected={selected.has(UNASSIGNED_ID)}
              onToggle={() => toggle(UNASSIGNED_ID)}
              onOnly={() => onChange([UNASSIGNED_ID])}
            >
              <UnassignedAvatar
                styles={`size-6
                ${selected.has(UNASSIGNED_ID) ? "" : "opacity-40"}`}
              />
              <span
                className={
                  selected.has(UNASSIGNED_ID) ? "text-fg" : "text-fg-muted"
                }
              >
                Unassigned
              </span>
            </FilterRow>
          )}
        </ul>
      </PopoverContent>
    </Popover>
  );
}
