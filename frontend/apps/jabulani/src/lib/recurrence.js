export function getDefaultRecurrenceEnd(startDate, frequency) {
  const endDate = new Date(startDate);
  const originalDay = endDate.getDate();
  const monthsToAdd = frequency === "monthly" ? 3 : 1;

  endDate.setDate(1);
  endDate.setMonth(endDate.getMonth() + monthsToAdd);

  const lastDayOfTargetMonth = new Date(
    endDate.getFullYear(),
    endDate.getMonth() + 1,
    0
  ).getDate();
  endDate.setDate(Math.min(originalDay, lastDayOfTargetMonth));

  return endDate;
}
