import defaultConfig from "./src/tailwind.config.js"
export default {
	...defaultConfig,
	content:[
		"src/components/**/*.{js,ts,jsx,tsx,mdx}"
	]
}

