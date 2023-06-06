import "../../index.css"
import { TextInput } from "../../components/atoms/input.jsx";
import { Search } from "@jengaicons/react";


export default {
  title: 'Atoms/TextInput',
  component: TextInput,
  tags: ['autodocs'],
  argTypes: {},
};


export const DefaultTextInput = {
  args: {
    label: "Default",
    value: "Hello",
    className: 'w-full'
  }
}


export const PrefixIconTextInput = {
  args: {
    label: "Default",
    value: "Hello",
    prefix: Search
  }
}

export const PrefixTextTextInput = {
  args: {
    label: "Default",
    value: "Hello",
    prefix: "$"
  }
}

export const PostfixTextInput = {
  args: {
    label: "Default",
    value: "Hello",
    suffix: "lbs"
  }
}


export const ShowClearTextInput = {
  args: {
    label: "Default",
    value: "Hello",
    showclear: true,
  }
}

export const ErrorTextInput = {
  args: {
    label: "Default",
    value: "Hello",
    showclear: true,
    prefix: Search,
    error: true
  }
}

export const DisabledTextInput = {
  args: {
    label: "Default",
    value: "Hello",
    showclear: true,
    prefix: Search,
    disabled: true
  }
}