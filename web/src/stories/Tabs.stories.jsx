import "../index.css"
import {TabsX} from "../components/atoms/tabs";

export default {
  title: 'Atoms/Tabs',
  component: TabsX,
  tags: ['autodocs'],
}

export const PrimaryTabs = {
  args: {
    fitted: true,
    children: [
      // <Tab label="Projects"   />,
      // <Tab label="Clusters"  />,
      // <Tab label="Domains"  />,
      // <Tab label="Container Registry" />,
      // <Tab label="Settings"  />,
    ]
  }
}