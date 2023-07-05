import "~/lib/app-setup/index.css"
import { NavTabs } from "../../components/atoms/tabs";

export default {
  title: 'Atoms/Tabs',
  component: NavTabs,
  tags: ['autodocs'],
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