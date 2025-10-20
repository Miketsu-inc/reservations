function applyTheme() {
  if (
    localStorage.theme === "dark" ||
    (!("theme" in localStorage) &&
      window.matchMedia("(prefers-color-scheme: dark)").matches)
  ) {
    document.documentElement.classList.add("dark");
    document.documentElement.setAttribute("data-ag-theme-mode", "dark");
  } else {
    document.documentElement.classList.remove("dark");
    document.documentElement.setAttribute("data-ag-theme-mode", "light");
  }
}

// Toggle theme manually and save preference
// eslint-disable-next-line no-unused-vars
function toggleTheme() {
  if (document.documentElement.classList.contains("dark")) {
    document.documentElement.classList.remove("dark");
    document.documentElement.setAttribute("data-ag-theme-mode", "light");
    localStorage.setItem("theme", "light");
  } else {
    document.documentElement.classList.add("dark");
    document.documentElement.setAttribute("data-ag-theme-mode", "dark");
    localStorage.setItem("theme", "dark");
  }
}

// Watch for changes to system theme
window
  .matchMedia("(prefers-color-scheme: dark)")
  .addEventListener("change", applyTheme);

// Apply the theme when the page loads
applyTheme();
