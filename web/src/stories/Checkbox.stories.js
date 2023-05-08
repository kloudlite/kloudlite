import "../index.css"
import { Checkbox } from "../components/atoms/checkbox";

export default {
  title: 'Atoms/Checkbox',
  component: Checkbox,
  tags: ['autodocs'],
  argTypes: {},
};

export const Checked = {
  args:{
    label: "Checked",
    value: true
  }
}

export const Disabled = {
  args:{
    label: "Disabled",
    value: true,
    disabled: true
  }
}
