import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import {App} from './app.jsx'

const basePath = import.meta.env.BASE_URL;
ReactDOM.hydrateRoot(
  document.getElementById('app'),
  <BrowserRouter basename={basePath}>
    <App />
  </BrowserRouter>,
)
