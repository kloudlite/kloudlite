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
    value:"lion",
    label: "Wild animals",
    items: [
      {label: "Tiger", value: "tiger", key:"1"},
      {label: "Lion", value: "lion", key:"2"},
      {label: "Zebra", value: "zebra", key:"3"},
      {label: "Giraffe", value: "giraffe", disabled: true, key:"4"},
    ]
  }
}