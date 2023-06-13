import { withRouter } from "storybook-addon-react-router-v6";
import { TopBar } from "../components/organisms/top-bar";

export default {
    title: "Organisms/Top bar",
    component: TopBar,
    decorators:[withRouter],
    tags: ['autodocs'],
    parameters:{
      reactRouter: {
        routePath: '/',
        // routeParams: { userId: '42' },
      }
    }
}


export const DefaultTopBar = {
    args:{

    }
}