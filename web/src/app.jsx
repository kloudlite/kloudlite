import { Route, Routes } from 'react-router-dom'
import appRoutes from '../routes.js'
import { NavBar } from "./components/header.jsx";

const pages = import.meta.glob('./pages/*', { eager: true })

export const App = ()=> {
  return (
    <div className={""}>
      <NavBar />
      <Routes>
        {
          appRoutes.map(({ path, page }) => {
            const RouteComp = pages[`./${page}`].default
            return <Route key={path} path={path} element={<RouteComp />}></Route>
          })
        }
      </Routes>
    </div>
  )
}
