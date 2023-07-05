import ReactDOMServer from 'react-dom/server'
import { StaticRouter } from 'react-router-dom/server'
import { App } from './app.jsx'
import css from "~/lib/app-setup/index.css?inline"

const basePath = import.meta.env.BASE_URL;

export function render(url, context) {
  return ReactDOMServer.renderToString(
    <StaticRouter location={url} context={context} basename={basePath}>
      <App />
    </StaticRouter>,
  )
}

export const cssString = css
