// Mirrors the actions and resources in backend/internal/permission
export const ACTIONS = {
  MERCHANT_DELETE: "merchant:delete",
  MERCHANT_RENAME: "merchant:rename",
  SETTINGS_EDIT: "settings:edit",
  LOCATION_IMAGES_EDIT: "location_images:edit",
  TEAM_MANAGE: "team:manage",
};

export const RESOURCES = {
  BOOKINGS: "bookings",
  BLOCKED_TIMES: "blocked_times",
  TEAM: "team",
};

export function createPermissionChecker(permissions, employeeId) {
  const actions = permissions?.actions ?? [];
  const view = permissions?.view ?? {};
  const edit = permissions?.edit ?? {};
  const manageableRoles = permissions?.manageable_roles ?? [];

  return {
    do: (action) => actions.includes(action),
    viewAll: (resource) => view[resource] === "all",
    // a record is visible when any of its employees is visible
    view: (resource, ...employeeIds) => {
      if (employeeIds.length === 0) return false;
      if (view[resource] === "all") return true;
      if (view[resource] === "own") return employeeIds.includes(employeeId);
      return false;
    },
    editAll: (resource) => edit[resource] === "all",
    // a record is editable only when every one of its employees is editable
    edit: (resource, ...employeeIds) => {
      if (employeeIds.length === 0) return false;
      if (edit[resource] === "all") return true;
      if (edit[resource] === "own") {
        return employeeIds.every((id) => id === employeeId);
      }
      return false;
    },
    manage: (role) => manageableRoles.includes(role),
  };
}
