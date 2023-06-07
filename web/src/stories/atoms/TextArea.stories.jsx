import { Button } from "../../components/atoms/button";
import { TextArea } from "../../components/atoms/input";
import "../../index.css"
import { Info, Search } from "@jengaicons/react";


export default {
  title: 'Atoms/Textarea',
  component: TextArea,
  tags: ['autodocs'],
  argTypes: {},
};


export const DefaultTextArea = {
  args: {
    label: "Default",
    value: "Hello",
    className: 'w-full'
  }
}

export const ErrorTextArea = {
  args: {
    label: "Default",
    value: "Hello",
    error: true,
    extra: <Button label="Link" style={'primary-plain'}/>,
    message: <span className="flex flex-row items-center gap-x-1"><Info size={16} color="currentColor"/> Required</span>
  }
}

export const DisabledTextArea = {
  args: {
    label: "Default",
    value: "Hello",
    disabled: true
  }
}