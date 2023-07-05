import { Link, Outlet } from "@remix-run/react"
import { useLocation, useMatch } from "@remix-run/react";
import { ActionList } from "../../../components/atoms/action-list.jsx"
import { SubHeader } from "../../../components/organisms/sub-header.jsx"

export default function ConsoleSettings() {

    const location = useLocation()
    console.log('location', location.pathname);

    let match = useMatch({
        path: "/:path/*"
    }, location.pathname)

    return <div className="flex flex-col gap-y-10">
        <SubHeader title={"Personal Account Settings"} />
        <div className="flex flex-row gap-x-25">
            <div className="w-45">
                <ActionList
                    LinkComponent={Link}
                    layoutId="settings"
                    value={match.params["*"]}
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
            </div>
            <Outlet />
        </div>
    </div>
}