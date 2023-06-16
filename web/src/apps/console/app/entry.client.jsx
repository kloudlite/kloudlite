import { RemixBrowser } from "@remix-run/react";
import { startTransition, StrictMode } from "react";
import { hydrateRoot } from "react-dom/client";
import { SSRProvider } from "react-aria"

hydrateRoot(
  document,
  <SSRProvider>
    <RemixBrowser />
  </SSRProvider>
);
