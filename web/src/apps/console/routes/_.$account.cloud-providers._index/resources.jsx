import {
  Cloud,
  DotsThreeVerticalFill,
  Info,
  PencilLine,
  Trash,
} from '@jengaicons/react';
import { dayjs } from '~/components/molecule/dayjs';
import { useState } from 'react';
import { Badge } from '~/components/atoms/badge';
import { IconButton } from '~/components/atoms/button';
import OptionList from '~/components/atoms/option-list';
import { cn } from '~/components/utils';
import {
  parseDisplaynameFromAnn,
  parseFromAnn,
  parseName,
  parseUpdationTime,
} from '~/console/server/r-urils/common';
import { keyconstants } from '~/console/server/r-urils/key-constants';

const ResourceItemExtraOptions = ({ open, setOpen, onEdit, onDelete }) => {
  return (
    <OptionList.Root open={open} onOpenChange={setOpen}>
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
        <OptionList.Separator />
        <OptionList.Item className="!text-text-critical" onSelect={onDelete}>
          <Trash size={16} />
          <span>Delete</span>
        </OptionList.Item>
      </OptionList.Content>
    </OptionList.Root>
  );
};

const Resources = ({ item, onEdit, onDelete, mode = 'list' }) => {
  const { name, id, cloudProviderName, status, lastupdated, author } = {
    name: parseDisplaynameFromAnn(item),
    id: parseName(item),
    cloudProviderName: item.cloudProviderName,
    status: 'running',
    author: parseFromAnn(item, keyconstants.author),
    lastupdated: (
      <span
        title={
          parseFromAnn(item, keyconstants.author)
            ? `Updated By ${parseFromAnn(
                item,
                keyconstants.author
              )}\nOn ${dayjs(parseUpdationTime(item)).format('LLL')}`
            : undefined
        }
      >
        {dayjs(parseUpdationTime(item)).fromNow()}
      </span>
    ),
  };

  const [openExtra, setOpenExtra] = useState(false);

  const ThumbnailComponent = () => (
    <span className="self-start">
      <Cloud size={20} />
    </span>
  );

  const TitleComponent = () => (
    <>
      <div className="flex flex-row gap-md items-center">
        <div className="headingMd text-text-default">{name}</div>
        {/* <div className="w-lg h-lg bg-icon-primary rounded-full" /> */}
      </div>
      <div className="bodyMd text-text-soft truncate">{id}</div>
    </>
  );

  const ClusterComponent = () => (
    <>
      <div className="w-[120px]">
        <Badge icon={<Info />}>{status}</Badge>
      </div>
      <div className="bodyMd text-text-strong w-[200px] flex flex-row items-center gap-lg">
        <Cloud size={14} />
        {cloudProviderName}
      </div>
    </>
  );

  const AuthorComponent = () => (
    <>
      <div className="bodyMd text-text-strong">{author}</div>
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
            <ThumbnailComponent />
            <div className="flex flex-col gap-sm w-[calc(100%-52px)] md:w-auto">
              {TitleComponent()}
            </div>
          </div>
          {OptionMenu()}
        </div>
        <div className="flex flex-col gap-md items-start">
          {ClusterComponent()}
        </div>
        <div className="flex flex-col items-start">{AuthorComponent()}</div>
      </div>
    );
  };

  const listView = () => (
    <>
      <div className="hidden md:flex flex-row items-center justify-between gap-3xl w-full">
        <div className="flex flex-1 flex-row items-center gap-xl">
          <ThumbnailComponent />
          <div className="flex flex-col gap-sm">{TitleComponent()}</div>
        </div>
        {ClusterComponent()}
        <div className="flex flex-col w-[200px]">{AuthorComponent()}</div>
        {OptionMenu()}
      </div>
      {gridView()}
    </>
  );

  if (mode === 'grid') return gridView();
  return listView();
};

export default Resources;
