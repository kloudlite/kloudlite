import "../../index.css"
import {NavTabs, NavTab} from "../../components/atoms/tabs";
import { withRouter } from "storybook-addon-react-router-v6";

export default {
  title: 'Atoms/Tabs',
  component: NavTabs,
  decorators:[withRouter],
  tags: ['autodocs'],
  parameters:{
    reactRouter: {
      routePath: '/',
      // routeParams: { userId: '42' },
    }
  }
}

export const PrimaryTabs = {
  args: {
    fitted: false,
    children: [
        <NavTab label="Projects" href="/" />
      ,
        <NavTab label="Clusters" href="/" />
      ,
        <NavTab label="Domains" href="/" />
      ,
        <NavTab label="Container Registry" href="/" />
      ,
    ]
  }
}