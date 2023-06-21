import { ArrowDownFill, CaretDownFill, CopySimple, Search } from "@jengaicons/react"
import { TextInput } from "../atoms/input"
import PropTypes from 'prop-types';
import { OptionList, OptionListGroup } from "../atoms/option";
import { useState } from "react";



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
            label={"menu"}
            selectedKeys={selected}
            items={[
                {
                    id: 'left',
                    children: [
                        {
                            id: 2, label: 'Final Copy (1)', icon: CopySimple, type: "checkbox",
                        },

                    ]
                },
                {
                    id: 'right',
                    children: [
                        { id: 3, label: 'index.ts', icon: CopySimple, type: "radio" },
                        { id: 4, label: 'package.json', icon: CopySimple, type: "label" },
                        { id: 1, label: 'license.txt', icon: CopySimple, type: "checkbox" }
                    ]
                }
            ]}
            searchFilter={true}
            onSearchFilterChange={(e) => {
                console.log(e);
            }}
            onSelectionChange={(e) => {
                console.log(e);
                // setSelected(Array.from(e.values()))
                if (selected.includes(e)) {
                    setSelected([...selected.filter((x) => x != e)])
                } else {
                    setSelected([...selected, e])
                }
            }}
        />

        {/* <OptionListGroup items={[
            {
                label: "menu1",
                type: "checkbox",
                disclosure: CaretDownFill,
                items: [
                    {
                        id: 'left',
                        children: [
                            {
                                id: 1, label: 'a', icon: CopySimple
                            }
                        ]
                    },
                    {
                        id: 'right',
                        children: [
                            { id: 2, label: 'b', icon: CopySimple },
                            { id: 3, label: 'c', icon: CopySimple },
                            { id: 4, label: 'd', icon: CopySimple }
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
                label: "menu12",
                type: "radio",
                items: [
                    {
                        id: 'left',
                        children: [
                            {
                                id: 1, label: 'a1', icon: CopySimple
                            }
                        ]
                    },
                    {
                        id: 'right',
                        children: [
                            { id: 2, label: 'a2', icon: CopySimple },
                            { id: 3, label: 'a3', icon: CopySimple },
                            { id: 4, label: 'a4', icon: CopySimple }
                        ]
                    }
                ],
                searchFilter: true,
                onSearchFilterChange: (e) => {
                    console.log(e);
                }
            }
        ]} /> */}
    </div>
}


Filters.propTypes = {
    onFilterTextChange: PropTypes.func,
    filterActions: PropTypes.element,
}
