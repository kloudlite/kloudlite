import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import mdx from '@mdx-js/rollup'
import {basePath} from "./src/base-path.js";

export default defineConfig({
  base:basePath,
  plugins: [react(), mdx()],
  build: {
    minify: true,
  },
})
