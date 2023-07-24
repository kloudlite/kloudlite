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
import { ToastProvider } from '~/components/molecule/toast';

export const links = () => [{ rel: 'stylesheet', href: stylesUrl }];

const EmptyWrapper = Fragment;

const NonIdleProgressBar = () => {
  const progress = useProgress();
  const { state } = useNavigation();
  useEffect(() => {
    if (state !== 'idle') {
      console.log(state);
      progress.show();
    } else if (progress.visible) {
      progress.hide();
    }
  }, [state]);
  return null;
};

const Root = ({ Wrapper = EmptyWrapper }) => {
  const { NODE_ENV, PORT = 443, IS_LOCAL } = useLoaderData();
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
          <LiveReload port={IS_LOCAL === 'true' ? 4000 + Number(PORT) : 443} />
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
            <ToastProvider>
              <Wrapper>
                <Outlet />
              </Wrapper>
            </ToastProvider>
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
      ? { PORT: Number(process.env.PORT), IS_LOCAL: process.env.IS_LOCAL }
      : {}),
  };
};

export default Root;
