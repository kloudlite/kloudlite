import "../../index.css"
import { RadioGroup } from "../../components/atoms/radio"

export default {
  title: 'Atoms/RadioGroup',
  component: RadioGroup,
  tags: ['autodocs'],
  argTypes: {},
}

export const DefaultRadioGroup = {
  args: {
    value:"item1",
    items: [
      {label: "Item 1", value: "item1", key:"1"},
      {label: "Item 2", value: "item2", key:"2"},
      {label: "Item 3", value: "item3", key:"3"},
      {label: "Item 4", value: "item4", disabled: true, key:"4"},
    ],
    className:"gap-4",
  }
}