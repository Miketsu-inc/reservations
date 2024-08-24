import { SCREEN_LG, SCREEN_MD, SCREEN_SM, SCREEN_XL } from "./constants";

export function getBreakPoint(windowSize) {
  if (windowSize <= SCREEN_SM) {
    return "sm";
  } else if (windowSize > SCREEN_SM && windowSize <= SCREEN_MD) {
    return "md";
  } else if (windowSize > SCREEN_MD && windowSize <= SCREEN_LG) {
    return "lg";
  } else if (windowSize > SCREEN_LG && windowSize <= SCREEN_XL) {
    return "xl";
  } else if (windowSize > SCREEN_XL) {
    return "2xl";
  }
}
