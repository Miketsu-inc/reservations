// Detect system theme changes and apply the appropriate class
function applySystemTheme() {
  const systemPrefersDark = window.matchMedia(
    "(prefers-color-scheme: dark)"
  ).matches;

  if (systemPrefersDark) {
    document.documentElement.classList.add("dark");
  } else {
    document.documentElement.classList.remove("dark");
  }
}

// Apply the theme based on user's preference or system preference
function applySavedTheme() {
  const savedTheme = localStorage.getItem("theme");
  if (savedTheme === "dark") {
    document.documentElement.classList.add("dark");
  } else if (savedTheme === "light") {
    document.documentElement.classList.remove("dark");
  } else {
    // If no preference is saved, fall back to the system preference
    applySystemTheme();
  }
}

// Toggle theme manually and save preference
// function toggleTheme() {
//   if (document.documentElement.classList.contains("dark")) {
//     document.documentElement.classList.remove("dark");
//     localStorage.setItem("theme", "light");
//   } else {
//     document.documentElement.classList.add("dark");
//     localStorage.setItem("theme", "dark");
//   }
// }

// Watch for changes to system theme
window
  .matchMedia("(prefers-color-scheme: dark)")
  .addEventListener("change", applySystemTheme);

// Apply the theme when the page loads
applySavedTheme();
