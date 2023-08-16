import { Fragment, useEffect } from 'react';
import {
  Links,
  LiveReload,
  Meta,
  Outlet,
  Scripts,
  useLoaderData,
  useNavigation,
  Link,
} from '@remix-run/react';
import stylesUrl from '~/design-system/index.css';
import ProgressContainer, {
  useProgress,
} from '~/components/atoms/progress-bar';
import reactToast from 'react-toastify/dist/ReactToastify.css';
import { ToastContainer } from '~/components/molecule/toast';
import { redirect } from '@remix-run/node';

export const links = () => [
  { rel: 'stylesheet', href: stylesUrl },
  { rel: 'stylesheet', href: reactToast },
];

const EmptyWrapper = Fragment;

export const _404Main = () => {
  return (
    <div className="text-[5vw] flex gap-[1vw] justify-center items-center min-h-screen">
      <div className="flex flex-col items-center">
        <span className="text-text-critical text-[10vw]">404</span>
        <span className="text-text-warning uppercase animate-pulse">
          page not found
        </span>
        <Link
          to="/"
          className="text-text-primary text-[1rem] hover:underline hover:text-text-strong transition-all underline"
        >
          Home Page
        </Link>
      </div>
    </div>
  );
};

export const meta = () => {
  return [{ title: '404 | Not found' }];
};

const NonIdleProgressBar = () => {
  const progress = useProgress();
  const { state } = useNavigation();
  useEffect(() => {
    if (state !== 'idle') {
      progress.show();
    } else if (progress.visible) {
      progress.hide();
    }
  }, [state]);
  return null;
};

const Root = ({ Wrapper = EmptyWrapper }) => {
  const { NODE_ENV, DEVELOPER, URL_SUFFIX, KL_BASE_URL } = useLoaderData();

  return (
    <html lang="en" className="bg-surface-basic-subdued text-text-default">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width,initial-scale=1" />
        <Links />
        <Meta />
      </head>
      <body className="antialiased">
        {/* <Loading progress={transition} /> */}
        <script
          // eslint-disable-next-line react/no-danger
          dangerouslySetInnerHTML={{
            __html: `
${KL_BASE_URL ? `window.KL_BASE_URL = ${`'${KL_BASE_URL}'`}` : ''}
${
  NODE_ENV === 'development'
    ? `window.DEVELOPER = ${`'${DEVELOPER}'`}`
    : `window.NODE_ENV = ${`'${NODE_ENV}'`}`
}
${URL_SUFFIX ? `window.URL_SUFFIX = ${`'${URL_SUFFIX}'`}` : ''}
               `,
          }}
        />
        <LiveReload port={443} />
        <ProgressContainer>
          <NonIdleProgressBar />
          <ToastContainer />
          <Wrapper>
            <Outlet />
          </Wrapper>
        </ProgressContainer>
        <Scripts />
      </body>
    </html>
  );
};

export const loader = (ctx) => {
  if (ctx?.request?.headers?.get('referer')) {
    return redirect(ctx.request.url);
  }

  const nodeEnv = process.env.NODE_ENV;
  return {
    NODE_ENV: nodeEnv,
    ...(nodeEnv === 'development'
      ? { PORT: Number(process.env.PORT), DEVELOPER: process.env.DEVELOPER }
      : {}),

    ...(process.env.URL_SUFFIX ? { URL_SUFFIX: process.env.URL_SUFFIX } : {}),
    ...(process.env.KL_BASE_URL
      ? { KL_BASE_URL: process.env.KL_BASE_URL }
      : {}),
  };
};

// params of shouldRevalidate
//   {
//   actionResult,
//   currentParams,
//   currentUrl,
//   defaultShouldRevalidate,
//   formAction,
//   formData,
//   formEncType,
//   formMethod,
//   nextParams,
//   nextUrl,
// }

export const shouldRevalidate = () => false;

export default Root;
