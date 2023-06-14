import fs from 'node:fs'
import path from 'node:path'
import process from 'node:process';
import {fileURLToPath} from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const toAbsolute = (p) => path.resolve(__dirname, p);

const template = fs.readFileSync(toAbsolute('./dist/static/index.html'), 'utf-8');
const { render, cssString } = await import('./dist/server/entry-server.js');
const routes = await import('./routes.js');
const basePath = process.env.BASE_PATH || '';


(async () => {
	// Generate CSS
	fs.writeFileSync(toAbsolute('./dist/static/assets/style.css'), cssString);
	for (const route of routes.default) {
		const { path } = route;
		const context = {}
		const appHtml = render(`${basePath}${path}`, context)

		const html = template.replace(`<!--app-html-->`, appHtml)
		.replace(`<!--style-->`, `<link rel="stylesheet" href="${basePath}/assets/style.css">`)

		const filePath = `dist/static${path === '/' ? '/index' : path}.html`
		fs.writeFileSync(toAbsolute(filePath), html)
		console.log('pre-rendered:', filePath)
	}
})()
