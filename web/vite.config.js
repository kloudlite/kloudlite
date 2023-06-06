import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import mdx from '@mdx-js/rollup'

export default defineConfig({
	base:process.env.BASE_PATH || '',
	plugins: [react(), mdx()],
	build: {
		minify: true,
	},
})
