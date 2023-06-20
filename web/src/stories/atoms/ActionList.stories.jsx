import "../../index.css"
import { ArrowsDownUp, Check } from "@jengaicons/react";
import { ActionList } from "../../components/atoms/action-list";
import { withRouter } from "storybook-addon-react-router-v6";
import { createRemixStub } from "@remix-run/testing/dist/create-remix-stub";

export default {
    title: "Atoms/ActionList",
    component: ActionList,
    tags: ['autodocs',],
    decorators: [
        (Story) => {
            const RemixStub = createRemixStub([
                {
                    path: '/',
                    element: <Story />,
                },
            ]);

            return <RemixStub />;
        },
    ],
    argTypes: {},
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