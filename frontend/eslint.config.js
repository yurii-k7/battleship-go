import js from "@eslint/js";
import globals from "globals";
import pluginReact from "eslint-plugin-react";
import tsParser from "@typescript-eslint/parser";
import tsPlugin from "@typescript-eslint/eslint-plugin";

export default [
  { files: ["**/*.{js,mjs,cjs,ts,tsx,jsx}"] },
  { 
    languageOptions: { 
      globals: {
        ...globals.browser,
        ...globals.jest
      },
      parser: tsParser,
      parserOptions: {
        ecmaVersion: "latest",
        sourceType: "module",
        ecmaFeatures: {
          jsx: true
        }
      }
    }
  },
  js.configs.recommended,
  {
    plugins: {
      "@typescript-eslint": tsPlugin
    },
    rules: {
      ...tsPlugin.configs.recommended.rules
    }
  },
  {
    ...pluginReact.configs.flat.recommended,
    settings: {
      react: {
        version: "detect"
      }
    }
  },
];
