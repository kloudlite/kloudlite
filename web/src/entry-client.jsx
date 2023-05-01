import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import {App} from './app'
import {basePath} from "./base-path.js";

ReactDOM.hydrateRoot(
  document.getElementById('app'),
  <BrowserRouter basename={basePath}>
    <App />
  </BrowserRouter>,
)
