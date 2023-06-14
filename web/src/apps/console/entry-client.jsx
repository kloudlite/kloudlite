import { createRoot } from "react-dom/client";
import { BrowserRouter } from 'react-router-dom'
import { TextInput } from "../../components/atoms/input.jsx";
import "../../index.css"
import Container from './pages/container.jsx';

const basePath = import.meta.env.BASE_URL;

const root = createRoot(document.getElementById("app"));

root.render(
  <BrowserRouter basename={"/"}>
    <Container />
  </BrowserRouter>
);
