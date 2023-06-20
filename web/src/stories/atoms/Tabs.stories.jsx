import "../../index.css"
import { NavTabs, NavTab } from "../../components/atoms/tabs";
import { createRemixStub } from "@remix-run/testing/dist/create-remix-stub";

export default {
  title: 'Atoms/Tabs',
  component: NavTabs,
  decorators: [
    (Story) => {
      const RemixStub = createRemixStub([
        {
          path: '/',
          element: <Story />,
        },
      ]);

      return <RemixStub />;
    },
  ],
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