/** @type {import('tailwindcss').Config} */
export const content = [
  "./internal/templates/**/*.{templ,html,go}",
  "./static/**/*.{js,html}",
  "./cmd/**/*.go",
  "./internal/**/*.{go,templ}",
];
export const theme = {
  extend: {
    fontFamily: {
      sans: [
        "Inter",
        "-apple-system",
        "BlinkMacSystemFont",
        "Segoe UI",
        "Roboto",
        "Oxygen",
        "Ubuntu",
        "Cantarell",
        "Open Sans",
        "sans-serif",
      ],
    },
  },
};
export const plugins = [require("@tailwindcss/forms")];
