import { ReactNode } from 'react';
import ExtendedFilledTab, { IExtendedFilledTab } from './extended-filled-tab';

interface IExtendedFilledTabWithContainer extends IExtendedFilledTab {
  children?: ReactNode;
}
const ExtendedFilledTabWithContainer = ({
  children,
  items,
  value,
  size,
  onChange,
}: IExtendedFilledTabWithContainer) => {
  return (
    <div className="flex flex-col rounded border border-border-default bg-surface-basic-default shadow-button overflow-hidden">
      <div
        className="py-xl px-3xl flex flex-row
         items-center bg-surface-basic-subdued"
      >
        <ExtendedFilledTab
          items={items}
          value={value}
          size={size}
          onChange={onChange}
        />
      </div>
      <div className="p-3xl">{children}</div>
    </div>
  );
};

export default ExtendedFilledTabWithContainer;
