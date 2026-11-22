// Mirrors the permission names in backend/internal/permission
export const PERMISSIONS = {
  MERCHANT_DELETE: "merchant:delete",
  MERCHANT_RENAME: "merchant:rename",
  SETTINGS_EDIT: "settings:edit",
  LOCATION_IMAGES_EDIT: "location_images:edit",
  TEAM_MANAGE: "team:manage",
  TEAM_VIEW_ALL: "team:view_all",
  CALENDAR_VIEW_ALL: "calendar:view_all",
  BOOKINGS_ALL: "bookings:all",
  BLOCKED_TIMES_ALL: "blocked_times:all",
  AVATARS_ALL: "avatars:all",
};

export function hasPermission(permissions, permission) {
  return permissions?.includes(permission) ?? false;
}
