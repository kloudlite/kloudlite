import "../../index.css"
import { ArrowsDownUp, Check } from "@jengaicons/react";
import { ActionList } from "../../components/atoms/action-list";
import { withRouter } from "storybook-addon-react-router-v6";

export default {
    title: "Atoms/ActionList",
    component: ActionList,
    tags: ['autodocs',],
    decorators: [withRouter],
    argTypes: {},
    parameters: {
        reactRouter: {
            routePath: '/',
            // routeParams: { userId: '42' },
        }
    }
}



export const DangerActionList = {
    args: {
        value: "general",
        layoutId: "danger",
        items: [
            {
                label: "General",
                value: "general",
                LeftIconComp: ArrowsDownUp,
                RightIconComp: Check,
                key: "1"
            },
            {
                label: "Invoices",
                value: "invoices",
                key: "2"
            },
            {
                label: "Billing",
                key: "3",
                value: "billing"
            },
            {
                label: "User Management",
                key: "4",
                value: "usermanagement"
            },
            {
                label: "Security and Privacy",
                key: "5",
                critical: true,
                value: "securityandprivacy",
            },

        ]
    }
}