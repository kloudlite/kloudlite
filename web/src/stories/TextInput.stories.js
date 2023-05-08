import "../index.css"
import {Input} from "../components/atoms/input.jsx";


export default {
  title: 'Atoms/Input',
  component: Input,
  tags: ['autodocs'],
  argTypes: {},
};


export const DefaultTextInput = {
  args: {
    label: "Default",
  }
}

export const DisabledTextInput = {
  args: {
    label: "Disabled",
  }
}

export const ErrorTextInput = {
  args: {
    label: "Error",
  }
}

export const SuccessTextInput = {
  args: {
    label: "Success",
  }
}

export const TextArea = {
  args: {
    Component: "textarea",
  }
}

export const NumberInput = {
  args: {
    type: "number",
  }
}

export const PasswordInput = {
  args: {
    type: "password",
  }
}

export const WithInfoContent = {
  args: {
    infoContent: "Info Content"
  }
}