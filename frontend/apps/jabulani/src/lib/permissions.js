// Mirrors the actions and resources in backend/internal/permission
export const ACTIONS = {
  DELETE_MERCHANT: "merchant:delete",
  RENAME_MERCHANT: "merchant:rename",
  EDIT_SETTINGS: "settings:edit",
  MANAGE_TEAM: "team:manage",
  MANAGE_CUSTOMERS: "customers:manage",
  MANAGE_SERVICES: "services:manage",
};

export const RESOURCES = {
  BOOKINGS: "bookings",
  BLOCKED_TIMES: "blocked_times",
  TEAM: "team",
  SERVICES: "services",
  REVIEWS: "reviews",
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
