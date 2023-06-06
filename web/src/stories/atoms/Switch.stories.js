import "../../index.css"
import { Switch } from "../../components/atoms/switch";

export default {
  title: 'Atoms/Switch',
  component: Switch,
  tags: ['autodocs'],
  argTypes: {},
};

export const On = {
  args:{
    label: "Checked",
    checked: true
  }
}

export const Off = {
  args:{
    label: "Disabled",
    checked: false
  }
}


