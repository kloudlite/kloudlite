import "../../index.css"
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

