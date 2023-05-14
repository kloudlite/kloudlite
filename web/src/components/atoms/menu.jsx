import classNames from 'classnames';
import { PropTypes } from 'prop-types';
import { useState, useRef, useEffect } from 'react';
import { BounceIt } from '../bounce-it';
import { AnimatePresence, motion } from 'framer-motion';
import { createPortal } from 'react-dom';

import {
  Menu as MenuComp,
  MenuArrow,
  MenuArrowTip,
  MenuContent,
  MenuContextTrigger,
  MenuItem,
  MenuItemGroup,
  MenuItemGroupLabel,
  MenuOptionItem,
  MenuPositioner,
  MenuSeparator,
  MenuTrigger,
  MenuTriggerItem,
  Portal,
} from '@ark-ui/react'


export const Menu = ({items, value, onChange, placeholder})=>{
  return (<MenuComp act>
    <MenuTrigger onSelect={console.log}>Open menu</MenuTrigger>
    <Portal>
      <MenuPositioner>
        <MenuContent className='bg-surface-default rounded'>
          <MenuItem id="edit" asChild>
            <div className='focus:bg-surface-secondary-default'>Hi</div>
          </MenuItem>
          <MenuItem id="delete">Delete</MenuItem>
          <MenuItem id="export">Export</MenuItem>
          <MenuItem id="duplicate">Duplicate</MenuItem>
        </MenuContent>
      </MenuPositioner>
    </Portal>
  </MenuComp>)
}

Menu.propTypes = {
  items: PropTypes.arrayOf(PropTypes.shape({
    label: PropTypes.string,
    value: PropTypes.string,
  })),
  value: PropTypes.string,
  placeholder: PropTypes.string,
  onChange: PropTypes.func,
}

Menu.defaultProps = {
  items: [],
  value: "",
  placeholder: "",
  onChange: ()=>{},
}