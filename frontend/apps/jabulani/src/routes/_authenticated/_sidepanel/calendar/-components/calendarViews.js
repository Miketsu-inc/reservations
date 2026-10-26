// Each range has a single person view and a view with a column per team
// member. Week and agenda get too wide for columns, so they merge everyone.
const dayRange = {
  value: "day",
  label: "Day",
  view: "timeGridDay",
  teamView: "resourceTimeGridDay",
};
const twoDayRange = {
  value: "2day",
  label: "2 days",
  view: "timeGrid2Day",
  teamView: "resourceTimeGrid2Day",
};
const weekRange = {
  value: "week",
  label: "Week",
  view: "timeGridWeek",
  teamView: "timeGridWeek",
};
const agendaRange = {
  value: "agenda",
  label: "Agenda",
  view: "listWeek",
  teamView: "listWeek",
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
  return allRanges.find(
    (range) => range.view === view || range.teamView === view
  )?.value;
}

export function resolveCalendarView(ranges, range, isTeam) {
  const option = ranges.find(({ value }) => value === range) ?? ranges[0];

  return isTeam ? option.teamView : option.view;
}
