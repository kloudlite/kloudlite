import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { TextInput } from "../../components/atoms/input.jsx";
import "../../index.css"

const basePath = import.meta.env.BASE_URL;
ReactDOM.hydrateRoot(
  document.getElementById('app'),
  <BrowserRouter basename={basePath}>
    <TextInput />
  </BrowserRouter>,
)
