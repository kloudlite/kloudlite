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
  useRouteError,
  isRouteErrorResponse,
} from '@remix-run/react';
import stylesUrl from '~/design-system/index.css';
import ProgressContainer, {
  useProgress,
} from '~/components/atoms/progress-bar';
import reactToast from 'react-toastify/dist/ReactToastify.css';
import { ToastContainer } from '~/components/molecule/toast';
import { redirect } from '@remix-run/node';
import skeletonCSS from 'react-loading-skeleton/dist/skeleton.css';
import { motion } from 'framer-motion';
import { TopBar } from '~/components/organisms/top-bar';
import { BrandLogo } from '~/components/branding/brand-logo';
import Container from '~/components/atoms/container';

export const links = () => [
  { rel: 'stylesheet', href: stylesUrl },
  { rel: 'stylesheet', href: reactToast },
  { rel: 'stylesheet', href: skeletonCSS },
];

const EmptyWrapper = Fragment;

export const ErrorWrapper = ({ children, message }) => {
  return (
    <html lang="en" className="bg-surface-basic-subdued text-text-default">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width,initial-scale=1" />
        <Links />
        <Meta />
      </head>
      <body className="antialiased">
        <TopBar logo={<BrandLogo detailed />} />
        <Container>
          <motion.pre
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ ease: 'anticipate' }}
          >
            <div className="flex flex-col max-h-[80vh] w-full bg-surface-basic-input border border-surface-basic-pressed on my-4xl rounded-md p-4xl gap-xl overflow-hidden">
              <div className="font-bold text-xl text-[#A71B1B]">{message}</div>
              <div className="flex overflow-scroll">
                <div className="bg-[#A71B1B] w-2xl" />
                <div className="overflow-auto max-h-full p-2xl flex-1 flex bg-[#EBEBEB] text-[#640C0C]">
                  {children}
                </div>
              </div>
            </div>
          </motion.pre>
        </Container>
        <Scripts />
      </body>
    </html>
  );
};

export function ErrorBoundary() {
  const error = useRouteError();

  if (isRouteErrorResponse(error)) {
    return (
      <ErrorWrapper message={`${error.status} ${error.statusText}`}>
        <code>
          {typeof error.data === 'string'
            ? error.data
            : JSON.stringify(error.data, null, 2)}
        </code>
      </ErrorWrapper>
    );
  }

  if (error instanceof Error) {
    return (
      <ErrorWrapper message={error.message}>
        <code>
          {typeof error.stack === 'string'
            ? error.stack
            : JSON.stringify(error.stack, null, 2)}
        </code>
      </ErrorWrapper>
    );
  }

  return <h1>Unknown Error</h1>;
}

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
