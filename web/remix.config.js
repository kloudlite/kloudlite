export default {
    appDirectory: `src/apps/${process.env.APP}/app`,
    assetsBuildDirectory: `public/${process.env.APP}/assets`,
    serverBuildDirectory: `public/${process.env.APP}/server`,
    serverDependenciesToBundle: ["@jengaicons/react"],
    publicPath: `/${process.env.APP}/assets/`,
    cacheDirectory: `public/${process.env.APP}/.cache`,
    devServerPort: Number(process.env.PORT) + 5000,
    tailwind: true,
    ignoredRouteFiles: ["**/.*"],
    // serverModuleFormat: "esm",
    future: {
        v2_routeConvention: true,
        v2_headers: true,
        v2_meta: true,
        v2_normalizeFormMethod: true,
        v2_errorBoundary: true
    },
}
