import "~/lib/app-setup/index.css"
import { Switch } from "../../components/atoms/switch";

export default {
  title: 'Atoms/Switch',
  component: Switch,
  tags: ['autodocs'],
  argTypes: {},
};

export const On = {
  args: {
    checked: true
  }
}

export const Off = {
  args: {
    checked: false,
    onChange: (e) => console.log(e)
  }
}

export const SwitchDisabled = {
  args: {
    checked: true,
    disabled: true
  }
}


