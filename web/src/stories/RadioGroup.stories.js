import "../index.css"
import { RadioGroup } from "../components/atoms/radio"

export default {
  title: 'Atoms/RadioGroup',
  component: RadioGroup,
  tags: ['autodocs'],
  argTypes: {},
}

export const DefaultRadioGroup = {
  args: {
    items: [
      {label: "Item 1", value: "item1"},
      {label: "Item 2", value: "item2"},
      {label: "Item 3", value: "item3"},
      {label: "Item 4", value: "item4", disabled: true},
    ],
    className:"gap-4",
  }
}