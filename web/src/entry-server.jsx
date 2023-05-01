import ReactDOMServer from 'react-dom/server'
import { StaticRouter } from 'react-router-dom/server'
import { App } from './app'
import css from "./index.css?inline"
import {basePath} from "./base-path.js";

export function render(url, context) {
  return ReactDOMServer.renderToString(
    <StaticRouter location={url} context={context} basename={basePath}>
      <App />
    </StaticRouter>,
  )
}

export const cssString = css
