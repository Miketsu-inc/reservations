import breezyTheme from "@fullcalendar/react/themes/breezy";

function omitKeys(object, shouldOmit) {
  return Object.fromEntries(
    Object.entries(object).filter(([key]) => !shouldOmit(key))
  );
}

// FullCalendar joins class options from every layer instead of replacing them,
// so the listWeek view (agenda) can only be restyled by dropping the theme's list classes.
export default {
  ...breezyTheme,
  optionDefaults: omitKeys(breezyTheme.optionDefaults, (name) =>
    name.startsWith("listDay")
  ),
  views: omitKeys(breezyTheme.views, (name) => name === "list"),
};
