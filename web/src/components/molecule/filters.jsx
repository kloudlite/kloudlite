import { Search } from "@jengaicons/react"
import { TextInput } from "../atoms/input"
import PropTypes from 'prop-types';
import { cloneElement } from "react"


export const Filters = ({ onFilterTextChange, filterActions }) => {
    return <div className="flex flex-row items-center gap-2 w-full flex-wrap">
        <TextInput placeholder={'Filters'} prefix={Search} onChange={onFilterTextChange} className={'flex-1 min-w-[288px]'} />
       
        {filterActions && filterActions.map((child, index) => {
            return cloneElement(child, {
                key: index
            })
        })}
    </div>
}



Filters.propTypes = {
    onFilterTextChange: PropTypes.func,
    filterActions: PropTypes.arrayOf(function(propValue, key, componentName, location, propFullName) {
        console.log(propValue[key].type.name);
        if (propValue[key] && (propValue[key].type.name !== "Button" && propValue[key].type.name !== "ButtonGroup" && propValue[key].type.name !== "IconButton")) {
          return new Error(
            'Invalid prop `' + propFullName + '` supplied to' +
            ' `' + componentName + '`. Validation failed.'
          );
        }
        
      })
}
