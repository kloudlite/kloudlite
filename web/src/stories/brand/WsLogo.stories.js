import "../../index.css"
import {Avatar} from "../../components/atoms/avatar.jsx";
import {BrandLogo} from "../../components/branding/brand-logo.jsx";
import {WorkspacesLogo} from "../../components/branding/workspace-logo.jsx";

export default {
  title: 'Branding/WorkspacesLogo',
  component: WorkspacesLogo,
  tags: ['autodocs'],
  argTypes: {},
};

export const BasicLogo = {
  args:{

  }
}

export const DetailedLogo = {
  args: {
    detailed: true
  },
};

