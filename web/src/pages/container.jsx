import { BellFill, CaretDownFill } from "@jengaicons/react";
import { Button, IconButton } from "../components/atoms/button";
import { BrandLogo } from "../components/branding/brand-logo"
import { TopBar } from "../components/organisms/top-bar"
import { Profile } from "../components/molecule/profile";
import classNames from "classnames";
import { EmptyState } from "../components/molecule/empty-state";
import { Navigate, Route, Routes, matchPath, useLocation } from "react-router-dom";
import Projects from "./projects";
import Cluster from "./cluster";
import { useState } from "react";
import Settings from "./settings";
import GeneralSettings from "./settings/general";
import BillingSettings from "./settings/billing";
import NewProject from "./new-project";

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
            {match?.params.path != "newproject" && <TopBar
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
                            <IconButton IconComp={BellFill} style="plain" />
                            <Profile name="Astroman" size={"small"} subtitle={null} />
                        </div>
                    </>
                }
            />}
            <div className={classNames("max-w-[1184px] m-auto",
                {
                    "pt-[95px]": fixedHeader
                })}>
                <Routes>
                    <Route path="projects" Component={Projects} />
                    <Route path="newproject" Component={NewProject} />
                    <Route path="cluster" Component={Cluster} />
                    <Route path="settings" Component={Settings}>
                        <Route index element={<Navigate to={"general"} />} />
                        <Route path="general" Component={GeneralSettings} />
                        <Route path="billing" Component={BillingSettings} />
                    </Route>
                </Routes>
            </div>

        </div>
    )
}

export default Container