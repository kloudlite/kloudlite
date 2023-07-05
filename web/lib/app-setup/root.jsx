import { Links, LiveReload, Outlet, Scripts } from "@remix-run/react";
import { SSRProvider } from "react-aria"
import stylesUrl from "./index.css";
import { GoogleReCaptchaProvider } from "react-google-recaptcha-v3";
export const links = () => [
  { rel: "stylesheet", href: stylesUrl },
];

export default () => {
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
        <SSRProvider>
          <Outlet />
        </SSRProvider>
      </GoogleReCaptchaProvider>
      <Scripts />
      </body>
      </html>
  );
}