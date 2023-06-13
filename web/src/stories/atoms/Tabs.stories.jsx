import "../../index.css"
import { NavTabs, NavTab } from "../../components/atoms/tabs";
import { withRouter } from "storybook-addon-react-router-v6";

export default {
  title: 'Atoms/Tabs',
  component: NavTabs,
  decorators: [withRouter],
  tags: ['autodocs'],
  parameters: {
    reactRouter: {
      routePath: '/',
      // routeParams: { userId: '42' },
    }
  }
}

export const PrimaryTabs = {
  args: {
    value: "projects",
    layoutId: "projects",
    onChange: (e) => { console.log(e); },
    items: [
      {
        label: "Projects",
        href: "#",
        key: "projects",
        value: "projects"
      },
      {
        label: "Cluster",
        href: "#",
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
        href: "#",
        key: "settings",
        value: "settings"
      },
    ]
  }
}