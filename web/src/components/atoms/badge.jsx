import React from 'react';
import PropTypes from 'prop-types';
import classnames from "classnames";
import {BounceIt} from "../bounce-it.jsx";



/**
 * Button component for user interaction
 */
export const Badge = ({label, disabled, selected, onChange, leftIcon, rightIcon}) => {
  const [selectedState, setSelectedState] = React.useState(selected)
  React.useEffect(()=> {
    onChange(selectedState);
  }, [selectedState])
  return (
    <BounceIt>
      <button className={classnames(
        "rounded-full border bodyMd flex gap-0.5 px-3 py-1 cursor-pointer transition-all",
        "focus-visible:ring-2 focus:ring-blue-400",
        {
          "text-grey-900":!selected && !disabled,
          "text-grey-400":disabled,
          "text-grey-50":selected,
        }, {
          "pointer-events-none":disabled,
        }, {
          "border-zinc-300": !selected && !disabled,
          "border-zinc-50":disabled,
          "border-primary-600":selected,
        },{
          "bg-zinc-50 hover:bg-zinc-200 active:bg-zinc-300": !selected && !disabled,
          "bg-zinc-50":disabled,
          "bg-primary-700":selected,
        })} onClick={()=> {
        setSelectedState(()=> {
          return !selectedState
        })
      }}>
        {leftIcon}
        {label}
        {rightIcon}
      </button>
    </BounceIt>
  );
};

Badge.propTypes = {
  label: PropTypes.string.isRequired,
  selected: PropTypes.bool,
  onChange: PropTypes.func,
  disabled: PropTypes.bool,
};

Badge.defaultProps = {
  label: "test",
  selected: false,
  onChange: () => {},
  disabled: false,
};
