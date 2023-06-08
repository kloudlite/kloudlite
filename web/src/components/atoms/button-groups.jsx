import React, { useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import { ButtonBase } from "./button.jsx";
import classnames from "classnames";

export const ButtonGroup = ({ items, size, value, onChange, selectable }) => {
  const [currentValue, setCurrentValue] = useState(value);
  useEffect(() => {
    if (onChange) onChange(currentValue);
  }, [currentValue]);

  return (
    <div className={classnames("flex flex-row")}>

      {items && items.map((child, index) => {

        const sharpRight = index < items.length - 1;
        const sharpLeft = index > 0;

        return <ButtonBase
          label={child.label}
          key={child.key}
          size={size}
          style={"basic"}
          sharpLeft={sharpLeft}
          sharpRight={sharpRight}
          selected={(child.value == currentValue) && selectable}
          className={classnames({ "-ml-px": (sharpLeft || sharpRight) })}
          IconComp={child.icon}
          iconOnly={!child.label && child.icon}
          DisclosureComp={child.label && child.disclosureComp}
          onClick={() => {
            setCurrentValue(child.value);
          }}
        />
      })}
    </div>
  );
};


ButtonGroup.propTypes = {
  items: PropTypes.arrayOf(PropTypes.shape({
    label: PropTypes.string,
    value: PropTypes.string.isRequired,
    icon: function (props, propName, componentName) {
      if ((props['label'] == undefined || props['label'] == "") && (props[propName] == undefined || typeof (props[propName]) != 'object')) {
        return new Error('Either label or icon is required!');
      }
    },
    disclosureComp: PropTypes.object,
    key: PropTypes.string
  })).isRequired,

  selectable: PropTypes.bool,
  onChange: PropTypes.func,
  value: PropTypes.string,
  size: PropTypes.oneOf(["small", "medium", "large"]),
};

ButtonGroup.defaultProps = {
  size: 'medium',
  onChange: (e) => { console.log(e) },
  selectable: false,
  items: [
    {
      label: "item 1",
      value: "item1",
      key: "item1",
    },
    {
      value: "item2",
      label: "item 2",
      key: "item2"
    },
    {
      label: "item 3",
      value: "item3",
      key: "item3"
    },
    {
      label: "item 4",
      value: "item4",
      key: "item4"
    },
  ]
};
