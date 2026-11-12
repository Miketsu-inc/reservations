import { Cancel01Icon } from "@hugeicons/core-free-icons";
import {
  Icon,
  MultiSelect,
  SearchInput,
  Select,
  Toggle,
  ToggleGroup,
} from "@reservations/components";
import { toAvatar } from "@reservations/lib";

const REPLY_OPTIONS = [
  { value: "all", label: "All reviews" },
  { value: "unreplied", label: "Needs reply" },
  { value: "replied", label: "Replied" },
];

const SORT_OPTIONS = [
  { value: "newest", label: "Newest first" },
  { value: "oldest", label: "Oldest first" },
  { value: "highest", label: "Highest rating" },
  { value: "lowest", label: "Lowest rating" },
  { value: "most_relevant", label: "Most relevant" },
];

const CLEARED = {
  stars: "all",
  hasReply: "all",
  serviceIds: [],
  employeeIds: [],
  search: "",
};

export default function ReviewFilters({
  filters,
  onChange,
  services = [],
  team = [],
  totalCount,
}) {
  const serviceOptions = services.map((svc) => ({
    value: svc.id,
    label: svc.name,
  }));

  const teamOptions = team.map((member) => {
    const { name, img, initials } = toAvatar(member);
    return { value: member.id, label: name, img, initials };
  });

  const serviceIds = filters.serviceIds ?? [];
  const employeeIds = filters.employeeIds ?? [];

  const activeFilters = [
    ...(filters.stars && filters.stars !== "all"
      ? [
          {
            id: "stars",
            label: `${filters.stars} ${Number(filters.stars) === 1 ? "star" : "stars"}`,
            onClear: () => onChange({ stars: "all" }),
          },
        ]
      : []),
    ...serviceIds.map((id) => ({
      id: `service-${id}`,
      label: services.find((s) => s.id === id)?.name,
      onClear: () =>
        onChange({ serviceIds: serviceIds.filter((x) => x !== id) }),
    })),
    ...employeeIds.map((id) => {
      const member = team.find((m) => m.id === id);
      return {
        id: `employee-${id}`,
        label: member ? toAvatar(member).name : `Old employee`,
        onClear: () =>
          onChange({ employeeIds: employeeIds.filter((x) => x !== id) }),
      };
    }),
  ];

  return (
    <div className="flex flex-col gap-3 text-sm">
      <div className="flex flex-col gap-3 sm:flex-row sm:justify-between">
        <ToggleGroup
          variant="segmented"
          styles="w-full sm:w-fit"
          value={filters.hasReply ?? "all"}
          onValueChange={(val) => val && onChange({ hasReply: val })}
        >
          {REPLY_OPTIONS.map((opt) => (
            <Toggle key={opt.value} value={opt.value} styles="sm:flex-none">
              {opt.label}
            </Toggle>
          ))}
        </ToggleGroup>

        <div className="w-full sm:w-44">
          <Select
            options={SORT_OPTIONS}
            value={filters.sort || "newest"}
            onSelect={(opt) => onChange({ sort: opt.value })}
            placeholder="Sort by"
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-12">
        <div className="col-span-2 md:col-span-6">
          <SearchInput
            placeholder="Search reviews"
            searchText={filters.search}
            onChange={(val) => onChange({ search: val })}
          />
        </div>

        <div className="col-span-1 md:col-span-3">
          <MultiSelect
            options={serviceOptions}
            values={serviceIds}
            onSelect={(values) => onChange({ serviceIds: values })}
            placeholder="Select services"
            displayText="service"
          />
        </div>

        <div className="col-span-1 md:col-span-3">
          <MultiSelect
            options={teamOptions}
            values={employeeIds}
            onSelect={(values) => onChange({ employeeIds: values })}
            placeholder="Select team members"
            displayText="team member"
          />
        </div>
      </div>

      {activeFilters.length > 0 && (
        <div className="flex flex-wrap items-center gap-2">
          {activeFilters.map((filter) => (
            <span
              key={filter.id}
              className="text-bg bg-fg inline-flex items-center gap-1.5
                rounded-full px-3 py-1.5 text-xs"
            >
              {filter.label}
              <button
                type="button"
                onClick={filter.onClear}
                aria-label={`Remove filter ${filter.label}`}
              >
                <Icon icon={Cancel01Icon} styles="size-3.5" />
              </button>
            </span>
          ))}

          {activeFilters.length > 1 && (
            <button
              type="button"
              onClick={() => onChange(CLEARED)}
              className="text-fg-muted hover:text-fg text-xs"
            >
              Clear all
            </button>
          )}
        </div>
      )}

      {totalCount != null && (
        <p className="text-fg-muted">
          {totalCount} {totalCount === 1 ? "review" : "reviews"}
        </p>
      )}
    </div>
  );
}
