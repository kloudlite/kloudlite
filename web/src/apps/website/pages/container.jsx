import { BellFill, CaretDownFill } from "@jengaicons/react";
import { Button, IconButton } from "~/root/src/stories/components/atoms/button.jsx";
import { BrandLogo } from "~/root/src/stories/components/branding/brand-logo.jsx"
import { TopBar } from "~/root/src/stories/components/organisms/top-bar.jsx"
import { Profile } from "~/root/src/stories/components/molecule/profile.jsx";
import classNames from "classnames";
import { Route, Routes, matchPath, useLocation } from "react-router-dom";
import Projects from "./projects.jsx";
import Cluster from "./cluster.jsx";
import Settings from "./settings/index.jsx";
import GeneralSettings from "./settings/general.jsx";
import BillingSettings from "./settings/billing.jsx";

const Container = ({ children }) => {
    let fixedHeader = true
    const location = useLocation()

    let match = matchPath({
        path: "/main/:path/:subpath"
    }, location.pathname)

    if (!match)
        match = matchPath({
            path: "/main/:path"
        }, location.pathname)

    console.log(match);
    return (
        <div>
            <TopBar
                fixed={fixedHeader}
                logo={
                    <BrandLogo detailed size={20} />
                }
                tab={{
                    value: match?.params.path,
                    fitted: true,
                    layoutId: "projects",
                    onChange: (e) => { console.log(e); },
                    items: [
                        {
                            label: "Projects",
                            href: "projects",
                            key: "projects",
                            value: "projects"
                        },
                        {
                            label: "Cluster",
                            href: "cluster",
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
                            href: "settings",
                            key: "settings",
                            value: "settings"
                        },
                    ]
                }}
                actions={
                    <>
                        <Button label={"Nuveo"} style={"basic"} DisclosureComp={CaretDownFill} />
                        <div className="h-[15px] w-px bg-border-default mx-4"></div>
                        <div className="flex flex-row gap-2 items-center justify-center">
                            <IconButton icon={BellFill} style="plain" />
                            <Profile name="Astroman" size={"small"} subtitle={null} />
                        </div>
                    </>
                }
            />
            <div className={classNames("max-w-[1184px] m-auto",
                {
                    "pt-[95px]": fixedHeader
                })}>
                <Routes>
                    <Route path="projects" Component={Projects} />
                    <Route path="cluster" Component={Cluster} />
                    <Route path="settings" Component={Settings}>
                        <Route Component={GeneralSettings} index />
                        <Route path="billing" Component={BillingSettings} />
                    </Route>
                </Routes>
            </div>

        </div>
    )
}

export default Container