import defaultConfig from "../../tailwind.config"
export default {
	...defaultConfig,
	content: [
		...defaultConfig.content,
		"./app/**/*.{js,jsx,ts,tsx}"
	],
}

