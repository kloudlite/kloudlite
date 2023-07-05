import "~/lib/app-setup/index.css"
import {NumberInput} from "../../components/atoms/input.jsx";


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
    disabled: false,
    onChange: (e)=>console.log(e)
  }
}