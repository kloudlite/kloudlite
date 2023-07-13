import Root, { links as baseLinks } from "~/lib/app-setup/root"
import authStylesUrl from "./styles/index.css";
import Container from "./pages/container";

export const links = () => {
    return [
        ...baseLinks(),
        { rel: "stylesheet", href: authStylesUrl }
    ]
}

const Layout = ({ children }) => {
    return (
        // <SSRProvider>
        <>
            {children}
        </>
        // </SSRProvider>
    )
}

export default ({ ...props }) => {
    return (
        <Root {...props} Wrapper={Layout} />
    )
};