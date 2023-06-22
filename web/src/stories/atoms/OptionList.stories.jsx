import "../../index.css"
import { OptionItemTypes, OptionList } from '../../components/atoms/option';
import { Button } from "../../components/atoms/button";
import { CopySimple } from "@jengaicons/react";

export default {
    title: 'Atoms/OptionList',
    component: OptionList,
    tags: ['autodocs'],
    argTypes: {},
}

export const DefaultOptionList = {
    args: {
        items: [
            {
                key: 1,
                trigger: <Button label="Fruits" />,
                items: [
                    {
                        id: 1,
                        children: [
                            { id: 1, label: 'Apple', type: OptionItemTypes.LABEL },
                            { id: 2, label: 'Banana', type: OptionItemTypes.LABEL },
                            { id: 3, label: 'Orange', type: OptionItemTypes.LABEL },
                            { id: 4, label: 'Grapes', type: OptionItemTypes.LABEL },
                        ]
                    }
                ],
                searchFilter: true,
                onSelectionChange: (e) => {
                    console.log(e);
                },
                onSearchFilterChange: (e) => {
                    console.log(e);
                }
            }
        ]
    }
}

export const CheckboxOptionList = {
    args: {
        items: [
            {
                key: 1,
                trigger: <Button label="Fruits" />,
                items: [
                    {
                        id: 1,
                        children: [
                            { id: 1, label: 'Apple', type: OptionItemTypes.CHECKBOX },
                            { id: 2, label: 'Banana', type: OptionItemTypes.CHECKBOX },
                            { id: 3, label: 'Orange', type: OptionItemTypes.CHECKBOX },
                            { id: 4, label: 'Grapes', type: OptionItemTypes.CHECKBOX },
                        ]
                    }
                ],
                searchFilter: true,
                onSelectionChange: (e) => {
                    console.log(e);
                },
                onSearchFilterChange: (e) => {
                    console.log(e);
                }
            }
        ]
    }
}

export const RadioOptionList = {
    args: {
        items: [
            {
                key: 1,
                trigger: <Button label="Fruits" />,
                items: [
                    {
                        id: 1,
                        children: [
                            { id: 1, label: 'Apple', type: OptionItemTypes.RADIO },
                            { id: 2, label: 'Banana', type: OptionItemTypes.RADIO },
                            { id: 3, label: 'Orange', type: OptionItemTypes.RADIO },
                            { id: 4, label: 'Grapes', type: OptionItemTypes.RADIO },
                        ]
                    }
                ],
                searchFilter: true,
                onSelectionChange: (e) => {
                    console.log(e);
                },
                onSearchFilterChange: (e) => {
                    console.log(e);
                }
            }
        ]
    }
}

export const IconOptionList = {
    args: {
        items: [
            {
                key: 1,
                trigger: <Button label="Fruits" />,
                items: [
                    {
                        id: 1,
                        children: [
                            { id: 1, label: 'Apple', type: OptionItemTypes.LABEL, icon: CopySimple },
                            { id: 2, label: 'Banana', type: OptionItemTypes.LABEL, icon: CopySimple },
                            { id: 3, label: 'Orange', type: OptionItemTypes.LABEL, icon: CopySimple },
                            { id: 4, label: 'Grapes', type: OptionItemTypes.LABEL, icon: CopySimple },
                        ]
                    }
                ],
                searchFilter: true,
                onSelectionChange: (e) => {
                    console.log(e);
                },
                onSearchFilterChange: (e) => {
                    console.log(e);
                }
            }
        ]
    }
}

//ID must be different for parent and child
export const SectionOptionList = {
    args: {
        items: [
            {
                key: 1,
                trigger: <Button label="Fruits" />,
                items: [
                    {
                        id: 11,
                        children: [
                            { id: 1, label: 'Apple', type: OptionItemTypes.LABEL, icon: CopySimple },
                            { id: 2, label: 'Banana', type: OptionItemTypes.LABEL, icon: CopySimple },
                            { id: 3, label: 'Orange', type: OptionItemTypes.LABEL, icon: CopySimple },
                            { id: 4, label: 'Grapes', type: OptionItemTypes.LABEL, icon: CopySimple },
                        ]
                    },
                    {
                        id: 12,
                        children: [
                            {
                                id: 5, label: 'Pineapple', icon: CopySimple, type: OptionItemTypes.CHECKBOX,
                            },

                        ]
                    },
                ],
                searchFilter: true,
                onSelectionChange: (e) => {
                    console.log(e);
                },
                onSearchFilterChange: (e) => {
                    console.log(e);
                }
            },
        ]
    }
}

export const GroupOptionList = {
    args: {
        items: [
            {
                key: 1,
                trigger: <Button label="Fruits" />,
                items: [
                    {
                        id: 11,
                        children: [
                            { id: 1, label: 'Apple', type: OptionItemTypes.LABEL, icon: CopySimple },
                            { id: 2, label: 'Banana', type: OptionItemTypes.LABEL, icon: CopySimple },
                            { id: 3, label: 'Orange', type: OptionItemTypes.LABEL, icon: CopySimple },
                            { id: 4, label: 'Grapes', type: OptionItemTypes.LABEL, icon: CopySimple },
                        ]
                    },
                    {
                        id: 12,
                        children: [
                            {
                                id: 5, label: 'Pineapple', icon: CopySimple, type: OptionItemTypes.CHECKBOX,
                            },

                        ]
                    },
                ],
                searchFilter: true,
                onSelectionChange: (e) => {
                    console.log(e);
                },
                onSearchFilterChange: (e) => {
                    console.log(e);
                }
            },
            {
                key: 2,
                trigger: <Button label="Vegetables" />,
                items: [
                    {
                        id: 1,
                        children: [
                            { id: 1, label: 'Cabbage', type: OptionItemTypes.LABEL, icon: CopySimple },
                            { id: 2, label: 'Tomato', type: OptionItemTypes.LABEL, icon: CopySimple },
                            { id: 3, label: 'Potato', type: OptionItemTypes.LABEL, icon: CopySimple },
                            { id: 4, label: 'Pumpkin', type: OptionItemTypes.LABEL, icon: CopySimple },
                        ]
                    }
                ],
                searchFilter: true,
                onSelectionChange: (e) => {
                    console.log(e);
                },
                onSearchFilterChange: (e) => {
                    console.log(e);
                }
            }
        ]
    }
}
