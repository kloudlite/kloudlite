import { Outlet } from "@remix-run/react"

export default function IndexRoute() {

    return (
        <div>

            hi
            <Outlet />
        </div>
    )
}