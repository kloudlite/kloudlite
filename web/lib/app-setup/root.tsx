import { HeadersFunction } from '@remix-run/node';
import {
  Link,
  Links,
  LiveReload,
  Meta,
  Outlet,
  Scripts,
  isRouteErrorResponse,
  useLoaderData,
  useNavigation,
  useRouteError,
} from '@remix-run/react';
import { motion } from 'framer-motion';
import rcSlide from 'rc-slider/assets/index.css';
import { ReactNode, useEffect } from 'react';
import skeletonCSS from 'react-loading-skeleton/dist/skeleton.css';
import styleReactPulsable from 'react-pulsable/index.css';
import reactToast from 'react-toastify/dist/ReactToastify.css';
import Container from '~/components/atoms/container';
import ProgressContainer, {
  useProgress,
} from '~/components/atoms/progress-bar';
// import { SelectPortalContainer } from '~/components/atoms/select';
import Tooltip from '~/components/atoms/tooltip';
import { BrandLogo } from '~/components/branding/brand-logo';
import { ToastContainer } from '~/components/molecule/toast';
import { TopBar } from '~/components/organisms/top-bar';
import styleZenerSelect from '@oshq/react-select/index.css';
import stylesUrl from '~/design-system/index.css';
import rcss from 'react-highlightjs-logs/dist/index.css';
import { isDev } from '../client/helpers/log';
import { getClientEnv, getServerEnv } from '../configs/base-url.cjs';

export const links = () => [
  { rel: 'stylesheet', href: stylesUrl },
  { rel: 'stylesheet', href: reactToast },
  { rel: 'stylesheet', href: skeletonCSS },
  { rel: 'stylesheet', href: rcSlide },
  { rel: 'stylesheet', href: styleReactPulsable },
  { rel: 'stylesheet', href: styleZenerSelect },
  { rel: 'stylesheet', href: rcss },
];

export const ErrorWrapper = ({ children, message }: any) => {
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
              <div
                className="font-bold text-xl text-[#A71B1B] truncate"
                title={message}
              >
                {message}
              </div>
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
      <ErrorWrapper message={error.name}>
        <code>
          {/* eslint-disable-next-line no-nested-ternary */}
          {isDev
            ? typeof error.stack === 'string'
              ? error.stack
              : JSON.stringify(error.stack, null, 2)
            : typeof error.stack === 'string'
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

const Root = ({
  Wrapper = ({ children }: { children: any }) => children,
}: {
  Wrapper: (prop: { children: ReactNode }) => JSX.Element;
}) => {
  const env = useLoaderData();

  return (
    <html lang="en" className="bg-surface-basic-subdued text-text-default">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width,initial-scale=1" />
        <Links />
        <Meta />
      </head>
      <body className="antialiased">
        <div
          id="loadOverlay"
          style={{
            backgroundColor: '#777',
            position: 'absolute',
            top: '0px',
            left: '0px',
            width: '100vw',
            height: '100vh',
            zIndex: '2000',
          }}
        />

        {/* <Loading progress={transition} /> */}
        <script
          // eslint-disable-next-line react/no-danger
          dangerouslySetInnerHTML={{
            __html: getClientEnv(env),
          }}
        />
        <LiveReload port={443} />
        <Tooltip.Provider>
          <ProgressContainer>
            <NonIdleProgressBar />
            <ToastContainer position="bottom-left" />
            <Wrapper>
              <Outlet />
            </Wrapper>
          </ProgressContainer>
        </Tooltip.Provider>
        <Scripts />
      </body>
    </html>
  );
};

export const loader = () => {
  return getServerEnv();
};

export const headers: HeadersFunction = ({
  actionHeaders,
  loaderHeaders,
  parentHeaders,
  errorHeaders,
}) => {
  console.log(loaderHeaders, actionHeaders, parentHeaders, errorHeaders);
  return {
    'X-Stretchy-Pants': 'its for fun',
    'Cache-Control': 'max-age=300, s-maxage=3600',
  };
};

export const shouldRevalidate = () => false;

export default Root;
