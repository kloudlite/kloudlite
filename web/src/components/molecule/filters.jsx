import { Search } from "@jengaicons/react"
import { TextInput } from "../atoms/input"
import PropTypes from 'prop-types';
import { useState } from "react";
import { OptionList, OptionListGroup } from "../atoms/dropdown";
import { Button } from "../atoms/button";
import { Toast, ToastProvider } from "./toast";
import ToggleGroup from "../atoms/togglegroup";




export const Filters = ({ onFilterTextChange, filterActions }) => {

    const [selected, setSelected] = useState("hi")
    const [helloWorld, setHelloworld] = useState(true)

    return <div className="flex flex-row items-center gap-2 w-full flex-wrap">
        <TextInput
            placeholder={'Filters'}
            prefix={Search}
            onChange={onFilterTextChange}
            className={'flex-1 min-w-32'}
        />
        {filterActions && filterActions}
        <ToggleGroup value={"left"}>
            <ToggleGroup.Button content={"Hello world"} value={"left"} />
            <ToggleGroup.Button content={"Hello world"} value={"right"} />
        </ToggleGroup>

        <OptionListGroup>
            <OptionList
                trigger={<Button variant="outline" content="Menu" />}
                filter
                onFilterChange={(e) => { console.log(e) }}
            >
                <OptionList.CheckboxItem checked={helloWorld} onValueChange={setHelloworld}>
                    Hello world
                </OptionList.CheckboxItem>
                <OptionList.Separator />
                <OptionList.RadioGroup value={selected} onValueChange={setSelected}>
                    <OptionList.RadioGroupItem value="hello">
                        Hello
                    </OptionList.RadioGroupItem>
                    <OptionList.RadioGroupItem value="hi">
                        Hi
                    </OptionList.RadioGroupItem>
                </OptionList.RadioGroup>
            </OptionList>
            <OptionList
                trigger={<Button variant="outline" content="Menu" />}
            >
                <OptionList.CheckboxItem>
                    Hello world
                </OptionList.CheckboxItem>
                <OptionList.Separator />
                <OptionList.RadioGroup value="hi">
                    <OptionList.RadioGroupItem value="hello">
                        Hello
                    </OptionList.RadioGroupItem>
                    <OptionList.RadioGroupItem value="hi">
                        Hi
                    </OptionList.RadioGroupItem>
                </OptionList.RadioGroup>
            </OptionList>
        </OptionListGroup>

        <OptionList
            trigger={<Button variant="outline" content="Menu" />}
        >
            <OptionList.CheckboxItem>
                Hello world
            </OptionList.CheckboxItem>
            <OptionList.Separator />
            <OptionList.RadioGroup value="hi">
                <OptionList.RadioGroupItem value="hello">
                    Hello
                </OptionList.RadioGroupItem>
                <OptionList.RadioGroupItem value="hi">
                    Hi
                </OptionList.RadioGroupItem>
            </OptionList.RadioGroup>
        </OptionList>

        <ToastProvider>
            <Toast show />
            <Toast show />
            <Toast show />
        </ToastProvider>
    </div>
}


Filters.propTypes = {
    onFilterTextChange: PropTypes.func,
    filterActions: PropTypes.element,
}
