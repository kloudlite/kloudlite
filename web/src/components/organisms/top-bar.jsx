import classNames from "classnames"
import PropTypes from "prop-types"
import { NavTabs } from "../atoms/tabs"
import { BrandLogo } from "../branding/brand-logo"
import { Button, IconButton } from "../atoms/button"
import { Profile } from "../molecule/profile"
import { BellFill, CaretDownFill } from "@jengaicons/react"

export const TopBar = ({ tab, actions, logo, fixed }) => {
  return (<div className={classNames("border-b border-border-default bg-surface-subdued px-2.5 z-10",
    {
      "fixed top-0 left-0 right-0": fixed
    })}>
    <div className="flex flex-col m-auto">
      <div className="flex flex-row items-center justify-between py-2">
        {logo && logo}
        <div className="flex flex-row items-center justify-center">
          {actions && actions}
        </div>
      </div>
      {tab && <NavTabs value={tab.value} layoutId={tab.layoutId} fitted={tab.fitted} items={tab.items} onChange={tab.onChange} />}
    </div>
  </div>)
}


TopBar.propTypes = {
  tab: PropTypes.object,
  actions: PropTypes.any,
  logo: PropTypes.any,
  fixed: PropTypes.bool
}

TopBar.defaultProps = {
  logo: <BrandLogo detailed size={20} />,
  actions: <>
    <Button label={"Nuveo"} variant={"basic"} DisclosureComp={CaretDownFill} />
    <div className="h-3.75 w-px bg-border-default mx-4"></div>
    <div className="flex flex-row gap-2 items-center justify-center">
      <IconButton IconComp={BellFill} variant="plain" />
      <Profile name="Astroman" size={"small"} subtitle={null} />
    </div>
  </>,
  tab: {
    value: "projects",
    fitted: true,
    layoutId: "projects",
    onChange: (e) => { console.log(e); },
    items: [
      {
        label: "Projects",
        href: "#",
        key: "projects",
        value: "projects"
      },
      {
        label: "Cluster",
        href: "#",
        key: "cluster",
        value: "cluster"
      },
      {
        label: "Cloud provider",
        href: "#",
        key: "cloudprovider",
        value: "cloudprovider"
      },
      {
        label: "Domains",
        href: "#",
        key: "domains",
        value: "domains"
      },
      {
        label: "Container registry",
        href: "#",
        value: "containerregistry",
        key: "containerregistry"
      },
      {
        label: "VPN",
        href: "#",
        key: "vpn",
        value: "vpn"
      },
      {
        label: "Settings",
        href: "#",
        key: "settings",
        value: "settings"
      },
    ]
  }
}