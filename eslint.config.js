import js from "@eslint/js";
import tanstackRouter from "@tanstack/eslint-plugin-router";
import react from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import globals from "globals";

export default [
  {
    ignores: ["frontend/dist/**", "frontend/src/routeTree.gen.*"],
  },
  js.configs.recommended,
  react.configs.flat.recommended,
  react.configs.flat["jsx-runtime"],
  reactHooks.configs["recommended-latest"],
  ...tanstackRouter.configs["flat/recommended"],
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
      },
      ecmaVersion: "latest",
      sourceType: "module",
    },
    settings: { react: { version: "19.1" } },
    rules: {
      "react/prop-types": "off",
      "react/no-unescaped-entities": 0,
      "react/jsx-no-target-blank": "off",
      "no-unused-vars": [
        "error",
        {
          varsIgnorePattern: "_",
          argsIgnorePattern: "_",
          caughtErrorsIgnorePattern: "_",
        },
      ],
    },
  },
];
