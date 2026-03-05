import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const apiPrefixes = [
  "/healthz",
  "/me",
  "/heroes",
  "/cards",
  "/deck",
  "/matches",
  "/auth",
];

export default defineConfig({
  plugins: [react()],
  server: {
    host: "0.0.0.0",
    port: 5173,
    proxy: Object.fromEntries(
      apiPrefixes.map((prefix) => [
        prefix,
        {
          target: "http://localhost:1234",
          changeOrigin: true,
        },
      ]),
    ),
  },
});
