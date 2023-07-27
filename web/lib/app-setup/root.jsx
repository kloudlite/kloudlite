import { Fragment, useEffect } from 'react';
import {
  Links,
  LiveReload,
  Outlet,
  Scripts,
  useLoaderData,
  useNavigation,
} from '@remix-run/react';
import stylesUrl from '~/design-system/index.css';
import { GoogleReCaptchaProvider } from 'react-google-recaptcha-v3';
import ProgressContainer, {
  useProgress,
} from '~/components/atoms/progress-bar';
import reactToast from 'react-toastify/dist/ReactToastify.css';
import { ToastContainer } from 'react-toastify';

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
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width,initial-scale=1" />
        <title>Kloudlite</title>
        <Links />
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
            <LiveReload port={443} />
          </>
        )}
        <GoogleReCaptchaProvider
          reCaptchaKey="6LdE1domAAAAAFnI8BHwyNqkI6yKPXB1by3PLcai"
          scriptProps={{
            async: false, // optional, default to false,
            defer: false, // optional, default to false
            appendTo: 'head', // optional, default to "head", can be "head" or "body",
            nonce: undefined, // optional, default undefined
          }}
          container={{
            // optional to render inside custom element
            element: 'captcha',
            parameters: {
              badge: '[inline|bottomright|bottomleft]', // optional, default undefined
              theme: 'dark', // optional, default undefined
            },
          }}
        >
          <ProgressContainer>
            <NonIdleProgressBar />
            <ToastContainer />
            <Wrapper>
              <Outlet />
            </Wrapper>
          </ProgressContainer>
        </GoogleReCaptchaProvider>
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

export default Root;
