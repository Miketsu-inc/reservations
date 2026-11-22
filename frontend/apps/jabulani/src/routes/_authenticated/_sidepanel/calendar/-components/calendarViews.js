import {
  Layout2ColumnIcon,
  Layout3ColumnIcon,
  SquareIcon,
} from "@hugeicons/core-free-icons";
import { RESOURCES } from "@reservations/jabulani/lib";

const dayRange = {
  value: "day",
  label: "Day",
  icon: SquareIcon,
  view: "resourceTimeGridDay",
  hasTeamColumns: true,
};
const twoDayRange = {
  value: "2day",
  label: "2 days",
  icon: Layout2ColumnIcon,
  view: "resourceTimeGrid2Day",
  hasTeamColumns: true,
};
const weekRange = {
  value: "week",
  label: "Week",
  icon: Layout3ColumnIcon,
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

export const timeGridViews = new Set([
  dayRange.view,
  twoDayRange.view,
  weekRange.view,
]);

export function getCalendarRanges(can, isMobile) {
  if (isMobile) return [agendaRange];
  if (!can.viewAll(RESOURCES.CALENDAR)) return [weekRange];

  return [dayRange, twoDayRange, weekRange];
}

export function getCalendarRange(view) {
  return allRanges.find((range) => range.view === view)?.value;
}

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
