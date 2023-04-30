import { Link, Route, Routes } from 'react-router-dom'
import appRoutes from '../routes.js'


const pages = import.meta.glob('./pages/*', { eager: true })

export function App() {
  return (
    <>
      <nav>
        <ul>
          <li>
            <Link to={"/"}>Home</Link>
            <Link to={"/env"}>Env</Link>
            <Link to={"/about"}>About</Link>
          </li>
        </ul>
      </nav>
      <Routes>
        {
          appRoutes.map(({ path, page }) => {
            const RouteComp = pages[`./${page}`].default
            return <Route key={path} path={path} element={<RouteComp />}></Route>
          })
        }
      </Routes>
    </>
  )
}
