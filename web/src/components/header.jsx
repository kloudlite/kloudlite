import {Link, matchRoutes, useLocation} from "react-router-dom";
import classnames from "classnames";

const HeaderLink = ({href, children}) => {
  const location = useLocation()
  const isActive = location.pathname === href
  return (
    <Link to={href} className={classnames({
      "border border-l-0 border-r-0 border-t-1 border-b-1": isActive,
    }, "px-1")}>
      {children}
    </Link>
  )
}

export const NavBar = () => {
  return (
    <div className={"flex flex-row justify-between p-4 border border-b-2"}>
      <div>Kloudlite Logo</div>
      <div className={"flex gap-x-8"}>
        <HeaderLink href={"/"}>Home</HeaderLink>
        <HeaderLink href={"/features"}>Features</HeaderLink>
        <HeaderLink href={"/pricing"}>Pricing</HeaderLink>
        <HeaderLink href={"#"}>Resources</HeaderLink>
        {/*<HeaderLink href={"/"}>Blog</HeaderLink>*/}
        {/*<HeaderLink href={"/"}>Support</HeaderLink>*/}
        <HeaderLink href={"/about"}>About Us</HeaderLink>
        <HeaderLink href={"#"}>Login / Sign Up</HeaderLink>
      </div>
    </div>
  )
}
