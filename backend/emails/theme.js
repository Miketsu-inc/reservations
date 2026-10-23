// Light-mode copy of the color tokens in frontend/input.css (:root + @theme).
// Emails can't use var(), so the values are inlined here; keep in sync by hand.
export const emailTheme = `
@theme {
  --color-canvas: oklch(0.959 0.004 286.3);
  --color-bg: oklch(0.983 0.003 286.4);
  --color-bg-raised: oklch(1 0 0);
  --color-bg-overlay: oklch(1 0 0);
  --color-bg-sunken: oklch(0.959 0.004 286.3);
  --color-bg-input: oklch(1 0 0);
  --color-bg-backdrop: oklch(0.182 0.018 284.4 / 0.4);
  --color-bg-inactive: oklch(0.915 0.008 286.2);

  --color-hover: oklch(0.203 0.039 282.3 / 0.05);
  --color-active: oklch(0.203 0.039 282.3 / 0.09);

  --color-fg: oklch(0.198 0.01 285.5);
  --color-fg-muted: oklch(0.472 0.019 285.6);
  --color-fg-disabled: oklch(0.77 0.014 286.1);
  --color-fg-on-primary: oklch(1 0 0);
  --color-fg-primary: oklch(0.484 0.185 277);
  --color-fg-success: oklch(0.51 0.114 156.5);
  --color-fg-warning: oklch(0.521 0.115 67.4);
  --color-fg-danger: oklch(0.527 0.189 21.6);
  --color-fg-info: oklch(0.5 0.148 255.5);

  --color-border: oklch(0.918 0.008 286.2);
  --color-border-subtle: oklch(0.944 0.005 286.3);
  --color-border-strong: oklch(0.857 0.012 286.1);
  --color-border-hover: oklch(0.74 0.016 286);
  --color-border-primary: oklch(0.54 0.184 278.3 / 0.35);
  --color-border-success: oklch(0.615 0.145 154.2 / 0.3);
  --color-border-warning: oklch(0.718 0.154 70 / 0.38);
  --color-border-danger: oklch(0.586 0.199 21 / 0.3);
  --color-border-info: oklch(0.597 0.167 254.9 / 0.3);

  --color-primary: oklch(0.54 0.184 278.3);
  --color-primary-hover: oklch(0.496 0.182 277.5);
  --color-primary-active: oklch(0.453 0.173 277);
  --color-primary-subtle: oklch(0.54 0.184 278.3 / 0.1);
  --color-primary-subtle-hover: oklch(0.54 0.184 278.3 / 0.16);
  --color-primary-focus: oklch(0.54 0.184 278.3 / 0.28);

  --color-success: oklch(0.615 0.145 154.2);
  --color-success-subtle: oklch(0.615 0.145 154.2 / 0.12);
  --color-warning: oklch(0.718 0.154 70);
  --color-warning-subtle: oklch(0.766 0.16 72.6 / 0.16);
  --color-danger: oklch(0.586 0.199 21);
  --color-danger-hover: oklch(0.532 0.19 21.9);
  --color-danger-subtle: oklch(0.586 0.199 21 / 0.09);
  --color-info: oklch(0.597 0.167 254.9);
  --color-info-subtle: oklch(0.597 0.167 254.9 / 0.11);
}
`;
