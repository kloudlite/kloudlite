import { Outlet } from "@remix-run/react"
import Container from "../pages/container"

export default () => {
    return <Container>
        <Outlet />
    </Container>
}