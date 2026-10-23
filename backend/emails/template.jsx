import React from "react";
import { parse } from "smol-toml";
import en from "./locales/emails.en.toml" with { type: "text" };
import hu from "./locales/emails.hu.toml" with { type: "text" };
import { previewData } from "./previewData";

void React;

// `make email` sets EMAIL_PREVIEW so the dev server shows real text and sample
// data. The export step emits the Go template syntax instead.
const isPreview = process.env.EMAIL_PREVIEW === "1";
const previewLang = process.env.EMAIL_LANG ?? "en";

// imported as text so the dev server hot reloads when a locale file changes
const locales = { en, hu };
const loadMessages = () => parse(locales[previewLang]);

function fillData(text) {
  return text.replace(
    /\{\{\s*\.(\w+)\s*\}\}/g,
    (_, field) => previewData[field] ?? `[${field}]`,
  );
}

// text from emails.<lang>.toml
export function t(key) {
  if (!isPreview) return `{{ T .Lang \`${key}\` . }}`;

  const [section, name] = key.split(".");
  const text = loadMessages()[section]?.[name];
  return text === undefined ? `[${key}]` : fillData(text);
}

// field of the data passed to the template
export function v(field) {
  if (!isPreview) return `{{ .${field} }}`;

  return previewData[field] ?? `[${field}]`;
}

// renders children only if the field is set
export function If({ field, children }) {
  if (!isPreview) {
    return (
      <>
        {`{{ if .${field} }}`}
        {children}
        {"{{ end }}"}
      </>
    );
  }

  return previewData[field] ? children : null;
}
