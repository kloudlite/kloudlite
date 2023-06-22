import { ArrowDownFill, CaretDownFill, CopySimple, Search } from "@jengaicons/react"
import { TextInput } from "../atoms/input"
import PropTypes from 'prop-types';
import { useState } from "react";
import { Chip, ChipTypes } from "../atoms/chip";
import { OptionItemTypes, OptionList } from "../atoms/option";
import { Button } from "../atoms/button";
import { Profile } from "./profile";



export const Filters = ({ onFilterTextChange, filterActions }) => {

    const [selected, setSelected] = useState([])

    return <div className="flex flex-row items-center gap-2 w-full flex-wrap">
        <TextInput
            placeholder={'Filters'}
            prefix={Search}
            onChange={onFilterTextChange}
            className={'flex-1 min-w-72'}
        />
        {filterActions && filterActions}

        <OptionList
            items={[
                {
                    key: "1",
                    trigger: <Button label={"menu"} style={'primary'} />,
                    items: [
                        {
                            id: 'left',
                            children: [
                                {
                                    id: 2, label: 'Final Copy (1)', icon: CopySimple, type: OptionItemTypes.CHECKBOX,
                                },

                            ]
                        },
                        {
                            id: 'right',
                            children: [
                                { id: 3, label: 'index.ts', icon: CopySimple, type: OptionItemTypes.LABEL },
                                { id: 4, label: 'package.json', icon: CopySimple, type: OptionItemTypes.RADIO },
                                { id: 1, label: 'license.txt', icon: CopySimple, type: OptionItemTypes.CHECKBOX }
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
                },
                {
                    key: "2",
                    trigger: <Button label={"menu"} style={'basic'} />,
                    items: [
                        {
                            id: 'left',
                            children: [
                                {
                                    id: 2, label: 'Final Copy (1)', icon: CopySimple, type: OptionItemTypes.CHECKBOX,
                                },

                            ]
                        },
                        {
                            id: 'right',
                            children: [
                                { id: 3, label: 'index.ts', icon: CopySimple, type: OptionItemTypes.LABEL },
                                { id: 4, label: 'package.json', icon: CopySimple, type: OptionItemTypes.RADIO },
                                { id: 1, label: 'license.txt', icon: CopySimple, type: OptionItemTypes.CHECKBOX }
                            ]
                        }
                    ],
                    searchFilter: true,
                    onSearchFilterChange: (e) => {
                        console.log(e);
                    }
                }
            ]} />

        <OptionList
            items={[
                {
                    key: "1",
                    trigger: <Profile name={"Astroman"} />,
                    items: [
                        {
                            id: 'left',
                            children: [
                                {
                                    id: 2, label: 'Final Copy (1)', icon: CopySimple, type: OptionItemTypes.CHECKBOX,
                                },

                            ]
                        },
                        {
                            id: 'right',
                            children: [
                                { id: 3, label: 'index.ts', icon: CopySimple, type: OptionItemTypes.LABEL },
                                { id: 4, label: 'package.json', icon: CopySimple, type: OptionItemTypes.RADIO },
                                { id: 1, label: 'license.txt', icon: CopySimple, type: OptionItemTypes.CHECKBOX }
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
            ]} />

    </div>
}


Filters.propTypes = {
    onFilterTextChange: PropTypes.func,
    filterActions: PropTypes.element,
}
