/** @type { import('@storybook/react-vite').StorybookConfig } */
const config = {
  stories: [
    "../atoms/**/*.@(js|jsx|ts|tsx|mdx)",
    "../molecules/**/*.@(js|jsx|ts|tsx|mdx)",
    "../brand/**/*.@(js|jsx|ts|tsx|mdx)"
  ],
  addons: ["@storybook/addon-links", "@storybook/addon-essentials", "@storybook/addon-interactions"],
  framework: {
    name: "@storybook/react-vite",
    options: {},
  },
  docs: {
    autodocs: "tag",
  },
};
export default config;
