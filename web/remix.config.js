import { flatRoutes } from 'remix-flat-routes';

/**
 * @type {import("@remix-run/dev").AppConfig}
 */
export default {
  appDirectory: `src/apps/${process.env.APP}`,
  assetsBuildDirectory: `public/${process.env.APP}/assets`,
  serverBuildPath: `public/${process.env.APP}/server/index.js`,
  serverDependenciesToBundle: 'all',
  // publicPath: `/${process.env.APP}/assets/public/`,
  cacheDirectory: `public/${process.env.APP}/.cache`,
  // devServerPort: Number(process.env.PORT) + 3000,
  tailwind: true,
  ignoredRouteFiles: ['**/.*'],
  serverModuleFormat: 'cjs',
  watchPaths: ['./src/design-system/**', 'lib/**'],
  future: {
    v2_routeConvention: true,
    v2_headers: true,
    v2_meta: true,
    v2_normalizeFormMethod: true,
    v2_errorBoundary: true,
    v2_dev: {
      port: Number(process.env.PORT) + 4000,
    },
  },
  routes: async (defineRoutes) => {
    return flatRoutes('routes', defineRoutes, {
      appDir: `src/apps/${process.env.APP}`,
    });
  },
};
