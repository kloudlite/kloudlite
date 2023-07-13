import React from "react";
import { Links, LiveReload, Outlet, Scripts } from "@remix-run/react";
import stylesUrl from "~/design-system/index.css";
import { GoogleReCaptchaProvider } from "react-google-recaptcha-v3";
export const links = () => [
  { rel: "stylesheet", href: stylesUrl },
];

const EmptyWrapper = React.Fragment;

export default ({ Wrapper = EmptyWrapper }) => {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta
          name="viewport"
          content="width=device-width,initial-scale=1"
        />
        <title>Remix: So great, it's funny!</title>
        <Links />
      </head>
      <body className="antialiased">
        <LiveReload />
        <GoogleReCaptchaProvider
          reCaptchaKey="6LdE1domAAAAAFnI8BHwyNqkI6yKPXB1by3PLcai"
          scriptProps={{
            async: false, // optional, default to false,
            defer: false, // optional, default to false
            appendTo: 'head', // optional, default to "head", can be "head" or "body",
            nonce: undefined // optional, default undefined
          }}
          container={{ // optional to render inside custom element
            element: "captcha",
            parameters: {
              badge: '[inline|bottomright|bottomleft]', // optional, default undefined
              theme: 'dark', // optional, default undefined
            }
          }}>
          <Wrapper>
            <Outlet />
          </Wrapper>
        </GoogleReCaptchaProvider>
        <Scripts />
      </body>
    </html>
  );
}