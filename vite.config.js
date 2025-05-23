import { sveltekit } from "@sveltejs/kit/vite";
import { nodePolyfills } from "vite-plugin-node-polyfills";

/** @type {import('vite').UserConfig} */
const config = {
  build: {
    target: 'es2021'
  },
  plugins: [
    sveltekit(),
    nodePolyfills({
      globals: {
        Buffer: true
      }
    })
  ],
  server: {
    fs: {
      allow: ["./fonts"]
    },
    host: "0.0.0.0",
    port: 12001,
    cors: true,
    headers: {
      "Access-Control-Allow-Origin": "*",
      "X-Frame-Options": "ALLOWALL"
    }
  }
};

export default config;
