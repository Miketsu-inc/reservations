// Every time grid is a resource view, so switching between team columns and a
// merged grid only changes the resources instead of remounting the view.
// Week and agenda get too wide for columns, so they always merge everyone.
const dayRange = {
  value: "day",
  label: "Day",
  view: "resourceTimeGridDay",
  hasTeamColumns: true,
};
const twoDayRange = {
  value: "2day",
  label: "2 days",
  view: "resourceTimeGrid2Day",
  hasTeamColumns: true,
};
const weekRange = {
  value: "week",
  label: "Week",
  view: "resourceTimeGridWeek",
  hasTeamColumns: false,
};
const agendaRange = {
  value: "agenda",
  label: "Agenda",
  view: "listWeek",
  hasTeamColumns: false,
};

const allRanges = [dayRange, twoDayRange, weekRange, agendaRange];

export function canViewTeamCalendar(role) {
  return role === "owner" || role === "admin";
}

export function getCalendarRanges(role, isMobile) {
  if (isMobile) return [agendaRange];
  if (!canViewTeamCalendar(role)) return [weekRange];

  return [dayRange, twoDayRange, weekRange];
}

export function getCalendarRange(view) {
  return allRanges.find((range) => range.view === view)?.value;
}

/**
 * Picks the first of the candidate range values that is available, falling
 * back to the first available range.
 */
export function resolveCalendarRange(ranges, ...candidates) {
  for (const candidate of candidates) {
    const range = ranges.find(({ value }) => value === candidate);
    if (range) return range.value;
  }

  return ranges[0].value;
}

function findRange(range) {
  return allRanges.find(({ value }) => value === range);
}

export function getCalendarView(range) {
  return findRange(range).view;
}

export function hasTeamColumns(range) {
  return findRange(range).hasTeamColumns;
}
