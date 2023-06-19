import { RemixBrowser } from "@remix-run/react";
import { startTransition, StrictMode } from "react";
import { hydrateRoot } from "react-dom/client";


hydrateRoot(
  document,
  <RemixBrowser />
);
