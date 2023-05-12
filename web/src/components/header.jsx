import {Link, useLocation} from "react-router-dom";
import classnames from "classnames";
import { useLink } from "react-aria";

const HeaderLink = (props) => {
  const {href, children} = props;
  const location = useLocation()
  const {linkProps} = useLink(props)
  const isActive = location.pathname === href
  return (
    <Link 
      {...linkProps}
      to={href} className={classnames("flex transition-all hover:text-text-default font-medium headingSm items-center",{
      "text-text-default": isActive,
      "text-text-soft": !isActive,
    }, "px-1")}>
      {children}
    </Link>
  )
}

export const NavBar = () => {
  const logoLinkProps = useLink({href:"/"})
  return (
      <div className={"flex flex-row justify-between p-4"}>
        <a className="p-1" {...logoLinkProps} href="/">
          Kloudlite Draft
        </a>
        <div className={"flex gap-x-8"}>
          <HeaderLink href={"/"}>Home</HeaderLink>
          <HeaderLink href={"/features"}>Features</HeaderLink>
          <HeaderLink href={"/pricing"}>Pricing</HeaderLink>
          <HeaderLink href={"/resources"}>Resources</HeaderLink>
          {/*<HeaderLink href={"/"}>Blog</HeaderLink>*/}
          {/*<HeaderLink href={"/"}>Support</HeaderLink>*/}
          <HeaderLink href={"/about"}>About Us</HeaderLink>
          <HeaderLink href={"#"}>Login / Sign Up</HeaderLink>
        </div>
      </div>
  )
}

