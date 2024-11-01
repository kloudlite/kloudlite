import { HeadersFunction, LinksFunction } from '@remix-run/node';
import {
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
import rcSlide from 'rc-slider/assets/index.css';
import { ReactNode, useEffect } from 'react';
import rcss from 'react-highlightjs-logs/dist/index.css';
import rootCss from '@kloudlite/design-system/index.css';
import skeletonCSS from 'react-loading-skeleton/dist/skeleton.css';
import styleReactPulsable from 'react-pulsable/index.css';
import reactToast from 'react-toastify/dist/ReactToastify.css';
import { Button } from '@kloudlite/design-system/atoms/button';
import Container from '@kloudlite/design-system/atoms/container';
import ProgressContainer, {
  useProgress,
} from '@kloudlite/design-system/atoms/progress-bar';
import Tooltip from '@kloudlite/design-system/atoms/tooltip';
import { TooltipContainer } from '@kloudlite/design-system/atoms/tooltipV2';
import { BrandLogo } from '@kloudlite/design-system/branding/brand-logo';
import { ToastContainer } from '@kloudlite/design-system/molecule/toast';
import Page404 from '@kloudlite/design-system/organisms/page-404';
import { TopBar } from '@kloudlite/design-system/organisms/top-bar';
import { ChildrenProps } from '@kloudlite/design-system/types';
import stylesUrl from '~/root/index.css';
import tailwindBase from '~/root/tailwind-base.js';
import { ReloadIndicator } from '~/lib/client/components/reload-indicator';
import { isDev } from '~/lib/client/helpers/log';
import { getClientEnv, getServerEnv } from '~/root/lib/configs/base-url.cjs';
import { useDataFromMatches } from '../client/hooks/use-custom-matches';

export const links: LinksFunction = () => [
  { rel: 'stylesheet', href: stylesUrl },
  { rel: 'stylesheet', href: rootCss },
  { rel: 'stylesheet', href: reactToast },
  { rel: 'stylesheet', href: skeletonCSS },
  { rel: 'stylesheet', href: rcSlide },
  { rel: 'stylesheet', href: styleReactPulsable },
  { rel: 'stylesheet', href: rcss },
  {
    rel: 'stylesheet',
    href: 'https://fonts.googleapis.com/css2?family=Familjen+Grotesk:ital,wght@0,500;0,600;0,700;1,400;1,500;1,600;1,700&display=swap',
  },
  { rel: 'stylesheet', href: 'https://rsms.me/inter/inter.css' },
];

export const Serror = ({ children, message }: ChildrenProps & any) => {
  return (
    <main className="antialiased">
      <TopBar logo={<BrandLogo detailed />} />
      <Container>
        <pre>
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
        </pre>
      </Container>
      <Scripts />
    </main>
  );
};

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
        <Serror message={message}>{children}</Serror>
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
  return <Page404 link="/" />;
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
  tagId,
}: {
  Wrapper: (prop: { children: ReactNode }) => JSX.Element;
  tagId?: string;
}) => {
  const env = useLoaderData();
  const error = useDataFromMatches('error', '');

  return (
    <html lang="en" className="bg-surface-basic-subdued text-text-default">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width,initial-scale=1" />
        <Links />
        <Meta />
        {/* <script */}
        {/*   async */}
        {/*   src="https://www.googletagmanager.com/gtag/js?id=G-9GFNBFM718" */}
        {/* /> */}
        {/* <script> */}
        {/*   window.dataLayer = window.dataLayer || []; */}
        {/*   function gtag(){dataLayer.push(arguments);} */}
        {/*   gtag('js', new Date()); */}

        {/*   gtag('config', 'G-9GFNBFM718'); */}
        {/* </script> */}
        <script
          async
          src={`https://www.googletagmanager.com/gtag/js?id=${tagId}`}
        />
        <script
          async
          id="gtag-init"
          // eslint-disable-next-line react/no-danger
          dangerouslySetInnerHTML={{
            __html: `
                window.dataLayer = window.dataLayer || [];
          function gtag(){dataLayer.push(arguments);}
          gtag('js', new Date());

          gtag('config', '${tagId}');
              `,
          }}
        />

        {/* <script */}
        {/*   // @ts-ignore */}
        {/*   // eslint-disable-next-line react/no-danger */}
        {/*   dangerouslySetInnerHTML={{ */}
        {/*     __html: ` */}
        {/*         try { */}
        {/*           if (localStorage.theme === 'dark' || (!('theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)) { */}
        {/*             document.querySelector('html').setAttribute('data-theme', 'dark') */}
        {/*           } else { */}
        {/*             document.querySelector('html').setAttribute('data-theme', 'light')                   */}
        {/*           } */}
        {/*         } catch (_) {} */}
        {/*       `, */}
        {/*   }} */}
        {/* /> */}
      </head>
      <body className="antialiased">
        <div
          id="loadOverlay"
          style={{
            backgroundColor: tailwindBase.theme.colors.surface.basic.default,
            position: 'absolute',
            top: '0px',
            left: '0px',
            width: '100vw',
            height: '100vh',
            zIndex: '2000',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
          }}
        >
          Loading...
        </div>

        {/* <Loading progress={transition} /> */}
        <script
          // eslint-disable-next-line react/no-danger
          dangerouslySetInnerHTML={{
            __html: getClientEnv(env),
          }}
        />
        <LiveReload port={443} />
        <ToastContainer position="bottom-left" />
        <ProgressContainer>
          <ReloadIndicator />
          <NonIdleProgressBar />
          {error ? (
            <div>
              <Serror message="Server Side Error">
                <code>{JSON.stringify(error, null, 2)}</code>
              </Serror>

              <div className="flex items-center justify-center">
                <Button to="/teams" content="Go To Home Page" />
              </div>
            </div>
          ) : (
            <Wrapper>
              <Tooltip.Provider>
                <Outlet />
              </Tooltip.Provider>
            </Wrapper>
          )}
          <TooltipContainer />
        </ProgressContainer>
        <Scripts />
      </body>
    </html>
  );
};

export const loader = () => {
  return getServerEnv();
};

export const headers: HeadersFunction = () => {
  // TODO: Remove this header if not needed
  return {
    'Cache-Control': 'max-age=300, s-maxage=3600',
  };
};

export const shouldRevalidate = () => false;

export default Root;
