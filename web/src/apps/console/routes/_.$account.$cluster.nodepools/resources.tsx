import {
  DotsThreeVerticalFill,
  PencilLine,
  StopCircle,
  Trash,
} from '@jengaicons/react';
import { useState } from 'react';
import { IconButton } from '~/components/atoms/button';
import OptionList from '~/components/atoms/option-list';
import Tooltip from '~/components/atoms/tooltip';
import { dayjs } from '~/components/molecule/dayjs';
import { cn } from '~/components/utils';
import { parseFromAnn } from '~/console/server/r-urils/common';
import { keyconstants } from '~/console/server/r-urils/key-constants';

const ResourceItemExtraOptions = ({
  open,
  setOpen,
  onEdit,
  onStop,
  onDelete,
}) => {
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <IconButton
          variant="plain"
          icon={<DotsThreeVerticalFill />}
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
        <OptionList.Item onSelect={onStop}>
          <StopCircle size={16} />
          <span>Stop</span>
        </OptionList.Item>
        <OptionList.Separator />
        <OptionList.Item className="!text-text-critical" onSelect={onDelete}>
          <Trash size={16} />
          <span>Delete</span>
        </OptionList.Item>
      </OptionList.Content>
    </OptionList.Root>
  );
};

// Project resouce item for grid and list mode
// mode param is passed from parent element
const Resources = ({ mode = '', item, onEdit, onDelete, onStop }) => {
  const {
    name,
    nodes,
    status,
    capacity,
    node_type: nodeType,
    provisionMode,
    lastupdated,
  } = {
    name: item.displayName,
    status: `status: TODO`,
    capacity: `TODO`,
    nodes: [],
    node_type: parseFromAnn(item, keyconstants.node_type),
    provisionMode: `TODO`,
    lastupdated: (
      <span
        title={
          parseFromAnn(item, keyconstants.author)
            ? `Updated By ${parseFromAnn(
                item,
                keyconstants.author
              )}\nOn ${dayjs(item.updateTime).format('LLL')}`
            : undefined
        }
      >
        {dayjs(item.updateTime).fromNow()}
      </span>
    ),
  };

  const [openExtra, setOpenExtra] = useState(false);
  const StartComponent = () => (
    <div className="flex flex-row gap-md items-center">
      <div className="headingMd text-text-default">{name}</div>
    </div>
  );

  const MiddleComponent = () => (
    <>
      <div className="bodyMd text-text-strong text-end w-[100px]">{status}</div>
      <div className="bodyMd text-text-strong text-end w-[120px]">
        {capacity}
      </div>
      <div className="bodyMd text-text-strong text-end w-[160px]">
        {nodeType}
      </div>
      <div className="bodyMd text-text-strong text-end w-[120px]">
        {provisionMode}
      </div>
    </>
  );

  const EndComponent = () => (
    <div className="bodyMd text-text-strong text-end">{lastupdated}</div>
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
    />
  );

  const NodeStatus = () => (
    <Tooltip.Provider>
      <div className="flex flex-row gap-xl">
        <div className="flex flex-row gap-lg">
          {nodes?.map((node) => (
            <Tooltip.Root
              content={
                <div className="flex flex-col bodySm-medium text-text-strong">
                  <span>
                    <span className="text-text-default">{node.name}: </span>
                    Error
                  </span>
                  <span>{node.ip}</span>
                </div>
              }
              key={node.id}
            >
              <div
                className={cn('w-2xl h-2xl', {
                  'bg-icon-success': node.status === 'running',
                  'bg-icon-warning': node.status === 'starting',
                  'bg-icon-critical': node.status === 'stopped',
                })}
              />
            </Tooltip.Root>
          ))}
        </div>
        <span className="bodySm text-text-soft">
          {nodes.filter((node) => node.status === 'running').length}/
          {nodes.length} ready
        </span>
      </div>
    </Tooltip.Provider>
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
      <div className="flex flex-col gap-3xl w-full">
        <div className="hidden md:flex flex-row items-center justify-between gap-3xl w-full">
          <div className="flex flex-1 flex-row items-center gap-xl">
            <div className="flex flex-col gap-sm">{StartComponent()}</div>
          </div>
          {MiddleComponent()}
          <div className="flex flex-col w-[140px]">{EndComponent()}</div>
          {OptionMenu()}
        </div>
        {NodeStatus()}
      </div>
      {gridView()}
    </>
  );

  if (mode === 'grid') return gridView();
  return listView();
};

export default Resources;
