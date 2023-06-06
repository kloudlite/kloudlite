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
    checked: true
  }
}

export const DisabledChecked = {
  args:{
    label: "Disabled",
    disabled: true,
    checked: true
  }
}

export const DisabledUnchecked = {
  args:{
    label: "Disabled",
    disabled: true,
    checked: false
  }
}

export const Indeterminate = {
  args:{
    label: "Disabled",
    disabled: true,
    checked: false,
    indeterminate: true
  }
}
