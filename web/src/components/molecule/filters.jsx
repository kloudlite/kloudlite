import { Search } from "@jengaicons/react"
import { TextInput } from "../atoms/input"
import PropTypes from 'prop-types';


export const Filters = ({ onFilterTextChange, filterActions }) => {
    return <div className="flex flex-row items-center gap-2 w-full flex-wrap">
        <TextInput
            placeholder={'Filters'}
            prefix={Search}
            onChange={onFilterTextChange}
            className={'flex-1 min-w-[288px]'}
        />
        {filterActions}
    </div>
}

Filters.propTypes = {
    onFilterTextChange: PropTypes.func,
    filterActions: PropTypes.element,
}
