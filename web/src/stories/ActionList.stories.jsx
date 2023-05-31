import { ArrowsDownUp, Check } from "@jengaicons/react";
import { ActionButton, ActionList } from "../components/atoms/actionlist";

export default {
    title: "Atoms/ActionList",
    component: ActionList,
    tags: ['autodocs',],
    argTypes: {}
}


export const DefaultActionList = {
    args: {
        children: [
            <ActionButton label="One" LeftIconComp={ArrowsDownUp} RightIconComp={Check}/>,
            <ActionButton label="two" disabled={true}/>,
            <ActionButton label="three" />
        ]
    }
}