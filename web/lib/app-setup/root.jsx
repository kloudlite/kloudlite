import { Fragment, useEffect } from 'react';
import {
  Links,
  LiveReload,
  Outlet,
  Scripts,
  useLoaderData,
  useTransition,
} from '@remix-run/react';
import stylesUrl from '~/design-system/index.css';
import { GoogleReCaptchaProvider } from 'react-google-recaptcha-v3';

export const links = () => [{ rel: 'stylesheet', href: stylesUrl }];

const EmptyWrapper = Fragment;

const Loading = ({ progress }) => {
  if (progress.state !== 'idle') {
    return <span>loading</span>;
  }
  return null;
};

const Root = ({ Wrapper = EmptyWrapper }) => {
  const { NODE_ENV } = useLoaderData();
  const transition = useTransition();
  useEffect(() => {
    console.log(transition.state);
  }, [transition]);
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width,initial-scale=1" />
        <title>Kloudlite</title>
        <Links />
      </head>
      <body className="antialiased">
        <Loading progress={transition} />
        {NODE_ENV === 'development' && <LiveReload port={443} />}
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
          <Wrapper>
            <Outlet />
          </Wrapper>
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
  };
};

export default Root;
