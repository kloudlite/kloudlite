import defaultConfig from "../../design-system/tailwind.config"
export default {
	...defaultConfig,
	content: [
		...defaultConfig.content,
		"./pages/**/*.{js,ts,jsx,tsx,mdx}",
	],
}

