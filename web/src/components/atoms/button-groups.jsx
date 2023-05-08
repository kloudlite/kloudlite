import React from 'react';
import PropTypes from 'prop-types';
import {Button} from "./button";
import classnames from "classnames";


export const ButtonGroup = ({items, size, fullWidth }) => {
  return (
    <div className={classnames("flex", {
      "w-full":fullWidth
    })}>
      {items.map((item, index) => {
        // if not first item then add property sharpRight and if not last item then add property sharpLeft
        const sharpRight = index < items.length - 1;
        const sharpLeft = index > 0;
        return <Button nobounce key={item.value} size={size} label={item.label} style={
          "basic"
        } sharpRight={sharpRight} sharpLeft={sharpLeft} onClick={()=>{item.onClick}} className={classnames({
          "flex-1":fullWidth
        })}/>
      })}
    </div>
  );
};


ButtonGroup.propTypes = {
  items: PropTypes.arrayOf(PropTypes.shape({
    label: PropTypes.string.isRequired,
    value: PropTypes.string.isRequired,
    onClick: PropTypes.func,
  })).isRequired,
  fullWidth: PropTypes.bool,
  style: PropTypes.oneOf(["primary", "secondary", "critical"]),
  size: PropTypes.oneOf(["small", "medium", "large"]),
};

ButtonGroup.defaultProps = {
  style: 'primary',
  size: 'medium',
  fullWidth:false,
  items: [
    {
      label: "test",
      value: "test",
    },
    {
      label: "test2",
      value: "test2",
    },
    {
      label: "test3",
      value: "test3",
    }
  ],
};
