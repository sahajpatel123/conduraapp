import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}",
    "./lib/**/*.{ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        ink: {
          950: "#050607",
          900: "#0b0d10",
          850: "#101419",
          800: "#151a20",
        },
        signal: {
          cyan: "#4de6ff",
          green: "#7cf6b5",
          amber: "#f5b95f",
          red: "#ff6d7a",
        },
      },
      boxShadow: {
        command: "0 24px 80px rgba(0, 0, 0, 0.38)",
        line: "0 0 0 1px rgba(255,255,255,0.08)",
      },
      fontFamily: {
        sans: [
          "Inter",
          "ui-sans-serif",
          "system-ui",
          "-apple-system",
          "BlinkMacSystemFont",
          "Segoe UI",
          "sans-serif",
        ],
        mono: [
          "SFMono-Regular",
          "ui-monospace",
          "Menlo",
          "Monaco",
          "Consolas",
          "monospace",
        ],
      },
    },
  },
  plugins: [],
};

export default config;
