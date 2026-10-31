import { ArrowLeft01Icon } from "@hugeicons/core-free-icons";
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

function getName(member) {
  return `${member.first_name} ${member.last_name}`;
}

function MemberAvatar({ member, styles }) {
  return (
    <Avatar
      img={member.avatar_url}
      initials={`${member.first_name[0]}${member.last_name[0]}`}
      alt={getName(member)}
      styles={styles}
    />
  );
}

function getTriggerText(selectedMembers, isEveryone, currentEmployeeId) {
  if (isEveryone) return "Everyone";
  if (selectedMembers.length > 1) return `${selectedMembers.length} members`;

  const [member] = selectedMembers;
  return member.id === currentEmployeeId ? "Only me" : getName(member);
}

/**
 * Picks whose bookings the calendar shows. At least one member always stays
 * selected, since an empty calendar has nothing to show or create on.
 */
export default function CalendarMemberFilter({
  members,
  selectedIds,
  onChange,
  currentEmployeeId,
}) {
  const [isOpen, setIsOpen] = useState(false);
  const listRef = useRef(null);

  const selected = new Set(selectedIds);
  const selectedMembers = members.filter((member) => selected.has(member.id));
  const isEveryone = selectedMembers.length === members.length;

  function toggleMember(memberId) {
    if (!selected.has(memberId)) {
      onChange([...selectedIds, memberId]);
      return;
    }

    if (selectedIds.length > 1) {
      onChange(selectedIds.filter((id) => id !== memberId));
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

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          aria-label="Shown team members"
          className="bg-bg-input border-border-strong
            hover:not-data-popup-open:border-border-hover
            focus-visible:border-primary focus-visible:ring-primary-focus
            data-popup-open:border-primary data-popup-open:ring-primary-focus
            flex w-full min-w-0 cursor-pointer items-center gap-2 rounded-lg
            border px-2 py-1.5 text-left text-sm outline-hidden
            transition-[border-color,box-shadow] duration-150
            focus-visible:ring-3 data-popup-open:ring-3 sm:w-48"
        >
          <AvatarStack
            avatars={selectedMembers.map((member) => toAvatar(member))}
          />
          <span className="text-fg min-w-0 flex-1 truncate">
            {getTriggerText(selectedMembers, isEveryone, currentEmployeeId)}
          </span>
          <Icon
            icon={ArrowLeft01Icon}
            styles={`text-fg-muted size-5 shrink-0 transition-transform
              ${isOpen ? "rotate-90" : "-rotate-90"}`}
          />
        </button>
      </PopoverTrigger>
      <PopoverContent styles="w-72! p-1!" align="end">
        <ul
          ref={listRef}
          onKeyDown={handleKeyDown}
          className="max-h-80 overflow-y-auto dark:scheme-dark"
        >
          <li>
            <button
              type="button"
              data-filter-item
              onClick={() => onChange(members.map((member) => member.id))}
              className="hover:bg-hover focus-visible:bg-hover flex w-full
                cursor-pointer items-center gap-3 rounded-md px-2 py-2 text-left
                font-medium outline-hidden"
            >
              <CheckBox
                checked={isEveryone}
                readOnly
                tabIndex={-1}
                styles="pointer-events-none"
              />
              Everyone
            </button>
          </li>
          <li aria-hidden="true" className="border-border my-1 border-t" />
          {members.map((member) => {
            const isSelected = selected.has(member.id);

            return (
              <li key={member.id} className="group relative flex items-center">
                <button
                  type="button"
                  data-filter-item
                  aria-pressed={isSelected}
                  onClick={() => toggleMember(member.id)}
                  className="hover:bg-hover focus-visible:bg-hover flex min-w-0
                    flex-1 cursor-pointer items-center gap-3 rounded-md px-2
                    py-1.5 pr-14 text-left outline-hidden"
                >
                  <CheckBox
                    checked={isSelected}
                    readOnly
                    tabIndex={-1}
                    styles="pointer-events-none"
                  />
                  <MemberAvatar member={member} styles="size-6! text-[10px]!" />
                  <span className="min-w-0 truncate">{getName(member)}</span>
                  {member.id === currentEmployeeId && (
                    <span className="text-fg-muted shrink-0 text-xs">You</span>
                  )}
                </button>
                <button
                  type="button"
                  onClick={() => onChange([member.id])}
                  className="text-fg-muted hover:text-fg hover:bg-bg-sunken
                    focus-visible:bg-bg-sunken absolute right-1 cursor-pointer
                    rounded-md px-2 py-1 text-xs font-medium opacity-0
                    outline-hidden group-hover:opacity-100
                    focus-visible:opacity-100 pointer-coarse:opacity-100"
                >
                  Only
                </button>
              </li>
            );
          })}
        </ul>
      </PopoverContent>
    </Popover>
  );
}
