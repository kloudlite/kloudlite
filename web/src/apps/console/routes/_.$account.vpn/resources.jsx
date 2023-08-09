import {
  ArrowClockwise,
  DotsThreeVerticalFill,
  PencilLine,
  QrCode,
  Trash,
  WireGuardlogoFill,
} from '@jengaicons/react';
import { useState } from 'react';
import { IconButton } from '~/components/atoms/button';
import OptionList from '~/components/atoms/option-list';
import { cn } from '~/components/utils';

const ResourceItemExtraOptions = ({
  open,
  setOpen,
  onEdit,
  onQR,
  onWireguard,
  onReset,
  onDelete,
}) => {
  return (
    <OptionList open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <IconButton
          variant="plain"
          icon={DotsThreeVerticalFill}
          selected={open}
          onClick={(e) => {
            e.stopPropagation();
          }}
          onMouseDown={(e) => {
            e.stopPropagation();
          }}
          onPointerDown={(e) => {
            e.stopPropagation();
          }}
        />
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.Item onSelect={onEdit}>
          <PencilLine size={16} />
          <span>Edit</span>
        </OptionList.Item>
        <OptionList.Item onSelect={onQR}>
          <QrCode size={16} />
          <span>Show QR Code</span>
        </OptionList.Item>
        <OptionList.Item onSelect={onWireguard}>
          <WireGuardlogoFill size={16} />
          <span>Show Wirguard Config</span>
        </OptionList.Item>
        <OptionList.Separator />
        <OptionList.Item onSelect={onReset}>
          <ArrowClockwise size={16} />
          <span>Reset Config</span>
        </OptionList.Item>
        <OptionList.Item className="!text-text-critical" onSelect={onDelete}>
          <Trash size={16} />
          <span>Delete</span>
        </OptionList.Item>
      </OptionList.Content>
    </OptionList>
  );
};

// Project resouce item for grid and list mode
// mode param is passed from parent element
const Resources = ({
  mode,
  item,
  onEdit,
  onDelete,
  onStop,
  onQR,
  onWireguard,
}) => {
  const { name, id, cluster, author, ip, activity, lastupdated } = item;

  const [openExtra, setOpenExtra] = useState(false);
  const StartComponent = () => (
    <>
      <div className="flex flex-row gap-md items-center">
        <div className="headingMd text-text-default">{name}</div>
      </div>
      <div className="bodyMd text-text-soft truncate">{id}</div>
    </>
  );

  const MiddleComponent = () => (
    <>
      <div className="bodyMd text-text-strong text-start w-[263px]">{ip}</div>
      <div className="bodyMd text-text-strong text-start w-[100px]">
        {cluster}
      </div>
      <div className="bodyMd text-text-strong text-start w-[100px]">
        {author}
      </div>
    </>
  );

  const EndComponent = () => (
    <>
      <div className="bodyMd text-text-strong">{activity}</div>
      <div className="bodyMd text-text-soft">{lastupdated}</div>
    </>
  );

  const OptionMenu = () => (
    <ResourceItemExtraOptions
      open={openExtra}
      setOpen={setOpenExtra}
      onEdit={() => {
        if (onEdit) onEdit(item);
      }}
      onDelete={() => {
        if (onDelete) onDelete(item);
      }}
      onStop={() => {
        if (onStop) onStop(item);
      }}
      onQR={() => {
        if (onQR) onQR(item);
      }}
      onWireguard={() => {
        if (onWireguard) onWireguard(item);
      }}
    />
  );

  const gridView = () => {
    return (
      <div
        className={cn('flex flex-col gap-3xl w-full', {
          'md:hidden': mode === 'list',
        })}
      >
        <div className="flex flex-row items-center justify-between gap-lg w-full">
          <div className="flex flex-row items-center gap-xl w-[calc(100%-44px)] md:w-auto">
            <div className="flex flex-col gap-sm w-[calc(100%-52px)] md:w-auto">
              {StartComponent()}
            </div>
          </div>
          {OptionMenu()}
        </div>
        <div className="flex flex-col gap-md items-start">
          {MiddleComponent()}
        </div>
        <div className="flex flex-col items-start">{EndComponent()}</div>
      </div>
    );
  };

  const listView = () => (
    <>
      <div className="hidden md:flex flex-row items-center justify-between gap-3xl w-full">
        <div className="flex flex-1 flex-row items-center gap-xl">
          <div className="flex flex-col gap-sm">{StartComponent()}</div>
        </div>
        {MiddleComponent()}
        <div className="flex flex-col w-[160px]">{EndComponent()}</div>
        {OptionMenu()}
      </div>
      {gridView()}
    </>
  );

  if (mode === 'grid') return gridView();
  return listView();
};

export default Resources;
