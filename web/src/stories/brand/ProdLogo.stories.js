import "../../index.css"
import {Avatar} from "../../components/atoms/avatar.jsx";
import {BrandLogo} from "../../components/branding/brand-logo.jsx";
import {ProdLogo} from "../../components/branding/prod-logo.jsx";

export default {
  title: 'Branding/ProdLogo',
  component: ProdLogo,
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

