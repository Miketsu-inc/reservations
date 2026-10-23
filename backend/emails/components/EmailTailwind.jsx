import React from "react";
import { Tailwind } from "react-email";
import { emailTheme } from "../theme";

void React;

export default function EmailTailwind({ children }) {
  return <Tailwind theme={emailTheme}>{children}</Tailwind>;
}
