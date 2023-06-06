import "../index.css"
import {NumberInput} from "../components/atoms/input.jsx";


export default {
  title: 'Atoms/NumberInput',
  component: NumberInput,
  tags: ['autodocs'],
  argTypes: {},
};


export const DefaultTextInput = {
  args: {
    label: "Default",
    defaultValue:4,
  }
}