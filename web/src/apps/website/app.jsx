import { Route, Routes } from 'react-router-dom'
import appRoutes from './routes.js'
import { NavBar } from "~/root/src/stories/components/header.jsx";

const pages = import.meta.glob('./pages/*', { eager: true })

const mdxComponents = {
  a:(props)=>{
    return <a {...props} />
  },
  blockquote:(props)=>{
    return <blockquote {...props} />
  },
  br:(props)=>{
    return <br {...props} />
  },
  code:(props)=>{
    return <code {...props}/>
  },
  em:(props)=>{
    return <em {...props}/>
  },
  h1:(props)=>{
    return <h1 {...props}/>
  },
  h2:(props)=>{
    return <h2 {...props}/>
  },
  h3:(props)=>{
    return <h3 {...props}/>
  },
  h4:(props)=>{
    return <h4 {...props}/>
  },
  h5:(props)=>{
    return <h5 {...props}/>
  },
  h6:(props)=>{
    return <h6 {...props}/>
  },
  hr:(props)=>{
    return <hr {...props}/>
  },
  img:(props)=>{
    return <img {...props}/>
  },
  li:(props)=>{
    return <li {...props}/>
  },
  ol:(props)=>{
    return <ol {...props}/>
  },
  p:(props)=>{
    return <p {...props}/>
  },
  pre:(props)=>{
    return <pre {...props}/>
  },
  strong:(props)=>{
    return <strong {...props}/>
  },
  ul:(props)=>{
    return <ul {...props}/>
  },
}

export const App = ()=> {
  return (
    <div className={""}>
      <NavBar />
      <Routes>
        {
          appRoutes.map(({ path, page }) => {
            const RouteComp = pages[`./${page}`].default
            return <Route key={path} path={path} element={
              <RouteComp components={mdxComponents} />
            } />
          })
        }
      </Routes>
    </div>
  )
}
