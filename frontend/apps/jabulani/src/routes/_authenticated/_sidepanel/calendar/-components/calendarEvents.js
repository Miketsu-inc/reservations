import { DEFAULT_SERVICE_COLOR, toAvatar } from "@reservations/lib";
import { UNASSIGNED_ID } from "./CalendarMemberFilter";

function getContrastColor(color) {
  const hex = color.replace("#", "");

  const red = parseInt(hex.substring(0, 2), 16);
  const green = parseInt(hex.substring(2, 4), 16);
  const blue = parseInt(hex.substring(4, 6), 16);

  const [r, g, b] = [red, green, blue].map((channel) => {
    const value = channel / 255;
    return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
  });
  const luminance = r * 0.2126 + g * 0.7152 + b * 0.0722;
  return luminance > 0.179 ? "#000000" : "#ffffff";
}

export function transformBusinessHours(businessHours) {
  if (!businessHours) return {};

  return Object.entries(businessHours).map(([day, times]) => ({
    daysOfWeek: [parseInt(day)],
    startTime: times.start_time.slice(0, 5),
    endTime: times.end_time.slice(0, 5),
  }));
}

// Bookings and blocked times of deleted or deactivated members have nobody on
// the team, so they go to the unassigned column instead of disappearing.
function getResourceIds(employeeIds, memberIds) {
  const assignedIds = employeeIds.filter((id) => memberIds.has(id));
  return assignedIds.length > 0 ? assignedIds : [UNASSIGNED_ID];
}

function isUnassigned(resourceIds) {
  return resourceIds[0] === UNASSIGNED_ID;
}

export function formatBookings(data, memberIds) {
  if (data === undefined) return;

  return data.map((booking) => {
    const resourceIds = getResourceIds([booking.employee_id], memberIds);
    let title = "Walk-in";
    const isGroup = booking.booking_type !== "appointment";

    if (isGroup) {
      title = `${booking.service_name} - ${booking.participants.length}/${booking.max_participants}`;
    } else if (booking.participants.length > 0) {
      const participant = booking.participants[0];
      title = toAvatar(participant).name;
    }

    const serviceColor = booking.service_color ?? DEFAULT_SERVICE_COLOR;

    return {
      id: booking.id,
      resourceIds,
      title: title,
      start: booking.from_date,
      end: booking.to_date,
      color: serviceColor,
      contrastColor: getContrastColor(serviceColor),
      durationEditable: false,
      startEditable: new Date(booking.to_date) > new Date() ? true : false,
      extendedProps: {
        // this is a number unlike the normal 'id' which get's converted to a string
        id: booking.id,
        event_type: "booking",
        booking_type: booking.booking_type,
        booking_status: booking.booking_status,
        is_recurring: booking.is_recurring,
        max_participants: booking.max_participants,
        participants: booking.participants,
        merchant_note: booking.merchant_note,
        service_name: booking.service_name,
        service_color: serviceColor,
        service_id: booking.service_id,
        employee_id: booking.employee_id,
        is_unassigned: isUnassigned(resourceIds),
        duration: booking.duration,
        phases: booking.phases ?? [],
        price: booking.price,
        price_type: booking.price_type,
      },
    };
  });
}

export function formatBlockedTimes(data, memberIds) {
  if (data === undefined) return;

  return data.map((blockedTime) => {
    const resourceIds = getResourceIds(blockedTime.employee_ids, memberIds);
    const start = blockedTime.is_all_day
      ? blockedTime.blocked_day
      : blockedTime.from_date;
    const end = blockedTime.is_all_day ? undefined : blockedTime.to_date;
    const editableUntil = blockedTime.is_all_day
      ? new Date(`${blockedTime.blocked_day}T23:59:59`)
      : new Date(blockedTime.to_date);

    return {
      id: blockedTime.id,
      resourceIds,
      title: `${blockedTime.name} ${blockedTime?.icon || ""}`,
      start: start,
      end: end,
      color: "#6b6b78",
      contrastColor: getContrastColor("#6b6b78"),
      durationEditable: !blockedTime.is_all_day,
      allDay: blockedTime.is_all_day,
      startEditable: editableUntil > new Date(),
      extendedProps: {
        id: blockedTime.id,
        type: "blocked",
        name: blockedTime.name,
        icon: blockedTime.icon,
        blocked_type_id: blockedTime.blocked_type_id,
        employee_ids: blockedTime.employee_ids,
        is_unassigned: isUnassigned(resourceIds),
        allDay: blockedTime.is_all_day,
      },
    };
  });
}

// Defined once, so the query keeps returning the same array between renders.
export function sortMembersByName(members) {
  return members.toSorted((a, b) =>
    toAvatar(a).name.localeCompare(toAvatar(b).name)
  );
}

export function formatResources(teamMembers, showUnassigned) {
  const resources = teamMembers.map((member) => {
    const { img, initials, name } = toAvatar(member);
    return {
      id: member.id,
      title: name,
      avatarUrl: img,
      initials,
      isUnassigned: false,
    };
  });

  if (showUnassigned) {
    resources.push({
      id: UNASSIGNED_ID,
      title: "Unassigned",
      isUnassigned: true,
    });
  }
  return resources;
}

/**
 * Dropping an event into another employee's column only moves its resource,
 * so mirror the change into the employee fields the side panels read.
 */
export function applyDroppedEmployee({ event, oldResource, newResource }) {
  if (!newResource) return;

  const oldEmployeeId = Number(oldResource.id);
  const newEmployeeId = Number(newResource.id);

  if (event.extendedProps.type === "blocked") {
    // An unassigned blocked time has no old member to replace.
    const employeeIds = event.extendedProps.employee_ids.filter(
      (employeeId) => employeeId !== oldEmployeeId
    );
    event.setExtendedProp("employee_ids", [
      ...new Set([...employeeIds, newEmployeeId]),
    ]);
    return;
  }

  event.setExtendedProp("employee_id", newEmployeeId);
}
