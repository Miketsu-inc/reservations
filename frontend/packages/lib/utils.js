import {
  SCREEN_2XL,
  SCREEN_LG,
  SCREEN_MD,
  SCREEN_SM,
  SCREEN_XL,
} from "./constants";

export function getBreakPoint(windowSize) {
  if (windowSize >= SCREEN_2XL) return "2xl";
  if (windowSize >= SCREEN_XL) return "xl";
  if (windowSize >= SCREEN_LG) return "lg";
  if (windowSize >= SCREEN_MD) return "md";
  if (windowSize >= SCREEN_SM) return "sm";
  return "sm";
}
