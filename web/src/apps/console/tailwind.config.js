import defaultConfig from "../../tailwind.config"
export default {
	...defaultConfig,
	content: [
		...defaultConfig.content,
		"./index.html",
		"./pages/**/*.{js,ts,jsx,tsx,mdx}",
	],
}

