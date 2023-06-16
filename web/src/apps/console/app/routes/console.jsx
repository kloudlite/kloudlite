import Container from "../pages/container"
import { Outlet, Link } from "@remix-run/react"
export default function Console() {
    return <Container>
        <Outlet />
    </Container>
}