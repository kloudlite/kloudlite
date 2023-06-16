import { BellFill, CaretDownFill } from "@jengaicons/react";
import { Links, LiveReload, Outlet, useLocation, useMatch } from "@remix-run/react";
import classNames from "classnames";
import { TopBar } from "../../../../components/organisms/top-bar";
import { BrandLogo } from "../../../../components/branding/brand-logo";
import { Button, IconButton } from "../../../../components/atoms/button";
import { Profile } from "../../../../components/molecule/profile";

const Container = ({ children }) => {
    let fixedHeader = true

    const location = useLocation()
    console.log('location', location.pathname);
    let match = useMatch({
        path: "/console/:path"
    }, location.pathname)
    return (
        <div className="px-2.5">
            {"" != "newproject" && <TopBar
                fixed={fixedHeader}
                logo={
                    <BrandLogo detailed size={20} />
                }
                tab={{
                    value: match.params.path,
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
                    "pt-23.75": fixedHeader && !("" == "newproject"),
                    "pt-15": "" === "newproject"
                })}>
                {children}
            </div>

        </div>
    )
}

export default Container