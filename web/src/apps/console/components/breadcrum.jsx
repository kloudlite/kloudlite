/* eslint-disable no-unused-vars */
import { forwardRef } from 'react';
import { Button as NativeButton } from '~/components/atoms/button';
import { _false } from '~/components/utils';

export const Breadcrum = ({ children }) => {
  return <div className="flex flex-row gap-md items-center">{children}</div>;
};

export const Button = _false
  ? (
      {
        content,
        size = '',
        icon = null,
        variant = '',
        disabled = false,
        ref = null,
        prefix = null,
        block = false,
        onClick = (_) => _,
        loading = false,
        suffix = null,
        type = 'button',
        href = '',
        LinkComponent = null,
        selected = false,
        onMouseDown = (_) => _,
        onMousePointer = (_) => _,
        onPointerDown = (_) => _,
        className = '',
      } = {
        content: null,
      }
    ) => null
  : _false ||
    forwardRef((props, ref) => {
      return (
        <div className="flex flex-row gap-md items-center">
          <div className="text-text-disabled bodySm">/</div>
          {/* @ts-ignore */}
          <NativeButton {...props} size="md" variant="plain" ref={ref} />
        </div>
      );
    });

export default Breadcrum;
