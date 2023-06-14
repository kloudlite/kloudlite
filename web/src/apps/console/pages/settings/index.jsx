import { Outlet, Route, Routes, matchPath, useLocation } from "react-router-dom"
import { ActionList } from "../../../../components/atoms/action-list"
import { SubHeader } from "../../../../components/organisms/sub-header"
import GeneralSettings from "./general"

const Settings = ({ }) => {
    const location = useLocation()

    let match = matchPath({
        path: "/main/:path/:subpath"
    }, location.pathname)

    console.log(match);
    return (
        <div className="flex flex-col gap-y-[40px]">
            <SubHeader title={"Personal Account Settings"} />
            <div className="flex flex-row gap-x-[100px]">
                <ActionList
                    layoutId="settings"
                    value={match?.params.subpath ? match.params.subpath : "general"}
                    items={[
                        {
                            label: "General",
                            value: "general",
                            key: "general",
                            href: "general"
                        },
                        {
                            label: "Billing",
                            value: "billing",
                            key: "billing",
                            href: "billing"
                        },
                        {
                            label: "Invoices",
                            value: "invoices",
                            key: "invoices",
                        },
                        {
                            label: "User management",
                            value: "usermanagement",
                            key: "usermanagement",
                        },
                        {
                            label: "Security & Privacy",
                            value: "securityandprivacy",
                            key: "securityandprivacy",
                        }
                    ]} />
                <Outlet />
            </div>
        </div>
    )
}

export default Settings