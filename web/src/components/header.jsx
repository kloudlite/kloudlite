import {Link, useLocation} from "react-router-dom";
import classnames from "classnames";

const HeaderLink = (props) => {
  const {href, children} = props;
  const location = useLocation();
  const isActive = location.pathname === href;
  return (
    <Link
      to={href}
      className={classnames("flex transition-all hover:text-text-default font-medium headingSm items-center",{
      "text-text-default": isActive,
      "text-text-soft": !isActive,
    }, "px-1")}>
      {children}
    </Link>
  )
}

export const NavBar = () => {
  return (
      <div className={"flex flex-row justify-between p-4"}>
        <Link className="p-1" to={"/"}>
          Kloudlite Draft
        </Link>
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

