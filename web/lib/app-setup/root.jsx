import { Fragment, useEffect } from 'react';
import {
  Links,
  LiveReload,
  Meta,
  Outlet,
  Scripts,
  useLoaderData,
  useNavigation,
} from '@remix-run/react';
import stylesUrl from '~/design-system/index.css';
import ProgressContainer, {
  useProgress,
} from '~/components/atoms/progress-bar';
import reactToast from 'react-toastify/dist/ReactToastify.css';
import { ToastContainer } from '~/components/molecule/toast';

export const links = () => [
  { rel: 'stylesheet', href: stylesUrl },
  { rel: 'stylesheet', href: reactToast },
];

const EmptyWrapper = Fragment;

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
  const { NODE_ENV, DEVELOPER } = useLoaderData();

  return (
    <html lang="en" className="bg-surface-basic-default text-text-default">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width,initial-scale=1" />
        <Links />
        <Meta />
      </head>
      <body className="antialiased">
        {/* <Loading progress={transition} /> */}
        {NODE_ENV === 'development' && (
          <>
            <script
              // eslint-disable-next-line react/no-danger
              dangerouslySetInnerHTML={{
                __html: `window.DEVELOPER = ${`'${DEVELOPER}'`}
              `,
              }}
            />
            {/* <LiveReload port={443} /> */}
          </>
        )}
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

export const loader = () => {
  const nodeEnv = process.env.NODE_ENV;
  return {
    NODE_ENV: nodeEnv,
    ...(nodeEnv === 'development'
      ? { PORT: Number(process.env.PORT), DEVELOPER: process.env.DEVELOPER }
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
