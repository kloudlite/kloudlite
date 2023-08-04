import { useEffect, useState } from 'react';
import { useLoaderData, useOutletContext } from '@remix-run/react';
import {
  ArrowDown,
  ArrowUp,
  ArrowsDownUp,
  CaretDownFill,
  Cloud,
  CopySimple,
  DotsThreeVerticalFill,
  Info,
  List,
  PencilLine,
  Plus,
  PlusFill,
  Search,
  SquaresFour,
  Trash,
} from '@jengaicons/react';
import { SubHeader } from '~/components/organisms/sub-header.jsx';
import { Button, IconButton } from '~/components/atoms/button.jsx';
import Toolbar from '~/components/atoms/toolbar';
import OptionList from '~/components/atoms/option-list';
import Pagination from '~/components/molecule/pagination';
import { AnimatePresence, motion } from 'framer-motion';
import * as Chips from '~/components/atoms/chips';
import { cn } from '~/components/utils';
import { ChipGroupPaddingTop } from '~/design-system/tailwind-base';
import { Badge } from '~/components/atoms/badge';
import * as Popup from '~/components/molecule/popup';
import * as AlertDialog from '~/components/molecule/alert-dialog';
import { PasswordInput, TextInput } from '~/components/atoms/input';
import * as SelectInput from '~/components/atoms/select';
import logger from '~/root/lib/client/helpers/log';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { toast } from '~/components/molecule/toast';
import { useReload } from '~/root/lib/client/helpers/reloader';
import { useLog } from '~/root/lib/client/hooks/use-log';
import { dayjs } from '~/components/molecule/dayjs';
import ResourceList from '../components/resource-list';
import { EmptyState } from '../components/empty-state';
import ScrollArea from '../components/scroll-area';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { dummyData } from '../dummy/data';
import { IdSelector, idTypes } from '../components/id-selector';
import { useAPIClient } from '../server/utils/api-provider';
import { getSecretRef } from '../server/r-urils/secret-ref';
import {
  getMetadata,
  getPagination,
  parseDisplaynameFromAnn,
  parseFromAnn,
  parseName,
  parseUpdationTime,
} from '../server/r-urils/common';
import { keyconstants } from '../server/r-urils/key-constants';

const ClusterToolbar = ({ viewMode, setViewMode }) => {
  const [statusOptionListOpen, setStatusOptionListOpen] = useState(false);
  const [clusterOptionListOpen, setClusterOptionListOpen] = useState(false);
  const [sortbyOptionListOpen, setSortybyOptionListOpen] = useState(false);
  return (
    <div>
      {/* Toolbar for md and up */}
      <div className="hidden md:flex">
        <Toolbar>
          <div className="w-full">
            <Toolbar.TextInput placeholder="Search" prefixIcon={Search} />
          </div>
          <Toolbar.ButtonGroup>
            <StatusOptionList
              open={statusOptionListOpen}
              setOpen={setStatusOptionListOpen}
            />
            <ProviderOptionList
              open={clusterOptionListOpen}
              setOpen={setClusterOptionListOpen}
            />
          </Toolbar.ButtonGroup>
          <SortbyOptionList
            open={sortbyOptionListOpen}
            setOpen={setSortybyOptionListOpen}
          />
          <ViewToggle mode={viewMode} onModeChange={setViewMode} />
        </Toolbar>
      </div>

      {/* Toolbar for mobile screen */}
      <div className="flex md:hidden">
        <Toolbar>
          <div className="flex-1">
            <Toolbar.TextInput placeholder="Search" prefixIcon={Search} />
          </div>
          <Toolbar.Button content="Add filters" prefix={Plus} variant="basic" />
          <SortbyOptionList
            open={sortbyOptionListOpen}
            setOpen={setSortybyOptionListOpen}
          />
        </Toolbar>
      </div>
    </div>
  );
};

const ClusterFilters = ({ appliedFilters, setAppliedFilters }) => {
  return (
    <AnimatePresence initial={false}>
      {appliedFilters.length > 0 && (
        <motion.div
          className={cn('flex flex-row gap-xl relative')}
          initial={{
            height: 0,
            opacity: 0,
            paddingTop: '0px',
            overflow: 'hidden',
          }}
          animate={{
            height: '46px',
            opacity: 1,
            paddingTop: ChipGroupPaddingTop,
          }}
          exit={{
            height: 0,
            opacity: 0,
            paddingTop: '0px',
            overflow: 'hidden',
          }}
          transition={{
            ease: 'linear',
          }}
          onAnimationStart={(e) => console.log(e)}
        >
          <ScrollArea className="flex-1">
            <Chips.ChipGroup
              onRemove={(c) =>
                setAppliedFilters(appliedFilters.filter((a) => a.id !== c))
              }
            >
              {appliedFilters.map((af) => {
                return <Chips.Chip {...af} key={af.id} item={af} />;
              })}
            </Chips.ChipGroup>
          </ScrollArea>
          {appliedFilters.length > 0 && (
            <div className="flex flex-row items-center justify-center">
              <Button
                content="Clear all"
                variant="primary-plain"
                onClick={() => {
                  setAppliedFilters([]);
                }}
              />
            </div>
          )}
        </motion.div>
      )}
    </AnimatePresence>
  );
};

// Button for toggling between grid and list view
const ViewToggle = ({ mode, onModeChange }) => {
  const [m, setM] = useState(mode);
  useEffect(() => {
    if (onModeChange) onModeChange(m);
  }, [m]);
  return (
    <Toolbar.ButtonGroup value={m} onValueChange={setM} selectable>
      <Toolbar.ButtonGroup.IconButton icon={List} value="list" />
      <Toolbar.ButtonGroup.IconButton icon={SquaresFour} value="grid" />
    </Toolbar.ButtonGroup>
  );
};

// Project resouce item for grid and list mode
// mode param is passed from parent element
export const ResourceItem = ({ mode, secret }) => {
  const [editProviderPopup, setEditProviderPopup] = useState(false);
  const { name, id, providerRegion, status, lastupdated, author } = {
    name: parseDisplaynameFromAnn(secret),
    id: parseName(secret),
    providerRegion: parseFromAnn(secret, keyconstants.provider),
    status: 'running',
    lastupdated: (
      <span
        title={
          parseFromAnn(secret, keyconstants.author)
            ? `Updated By ${parseFromAnn(secret, keyconstants.author)}`
            : undefined
        }
      >
        {dayjs(parseUpdationTime(secret)).fromNow()}
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
        <Badge label={status} icon={Info} />
      </div>
      <div className="bodyMd text-text-strong w-[200px] flex flex-row items-center gap-lg">
        <Cloud size={14} />
        {providerRegion}
      </div>
    </>
  );

  const AuthorComponent = () => (
    <>
      <div className="bodyMd text-text-strong">{author}</div>
      <div className="bodyMd text-text-soft">{lastupdated}</div>
    </>
  );

  const [deleteProviderPopup, setDeleteProviderPopup] = useState(false);

  const OptionMenu = () => (
    <ResourceItemExtraOptions
      open={openExtra}
      setOpen={setOpenExtra}
      onEdit={() => setEditProviderPopup(true)}
      onDelete={() => setDeleteProviderPopup(true)}
    />
  );

  const DeleteAlert = () => {
    return (
      <AlertDialog.DialogRoot
        show={deleteProviderPopup}
        onOpenChange={setDeleteProviderPopup}
      >
        <AlertDialog.Header>Delete Cloud Provider</AlertDialog.Header>
        <AlertDialog.Content>
          Are you sure you want to delete &apos;kloud-root-ca.crt&apos;.
        </AlertDialog.Content>
        <AlertDialog.Footer>
          <AlertDialog.Button variant="basic" content="Cancel" />
          <AlertDialog.Button variant="critical" content="Delete" />
        </AlertDialog.Footer>
      </AlertDialog.DialogRoot>
    );
  };

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

        {/* Popup dialog for editing cloud provider */}
        <UpdatePopUp
          {...{
            secret,
            editProviderPopup,
            setEditProviderPopup,
          }}
        />
        <DeleteAlert />
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

// OptionList for various actions
const StatusOptionList = ({ open, setOpen }) => {
  const [statuses, setStatuses] = useState([
    { checked: false, content: 'Verified', id: 'verified' },
    { checked: false, content: 'Un-Verified', id: 'unverified' },
  ]);
  return (
    <OptionList open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <Toolbar.ButtonGroup.Button
          content="Status"
          variant="basic"
          suffix={CaretDownFill}
        />
      </OptionList.Trigger>
      <OptionList.Content>
        {statuses.map((status) => (
          <OptionList.CheckboxItem
            key={status.id}
            checked={status.checked}
            onValueChange={(e) =>
              setStatuses(
                statuses.map((stat) => {
                  return stat.id === status.id ? { ...stat, checked: e } : stat;
                })
              )
            }
            onSelect={(e) => e.preventDefault()}
          >
            {status.content}
          </OptionList.CheckboxItem>
        ))}
      </OptionList.Content>
    </OptionList>
  );
};

const ProviderOptionList = ({ open, setOpen }) => {
  const [providers, setProviders] = useState(dummyData.providers);
  return (
    <OptionList open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <Toolbar.ButtonGroup.Button
          content="Provider"
          variant="basic"
          suffix={CaretDownFill}
        />
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.TextInput
          placeholder="Filter provider"
          prefixIcon={Search}
        />
        {providers.map((provider) => (
          <OptionList.Item
            key={provider.id}
            onSelect={(e) => e.preventDefault()}
          >
            <div className="flex flex-row gap-xl">
              <CopySimple size={16} />
              {provider.content}
            </div>
          </OptionList.Item>
        ))}
      </OptionList.Content>
    </OptionList>
  );
};

const SortbyOptionList = ({ open, setOpen }) => {
  const [sortbyProperty, setSortbyProperty] = useState('updated');
  const [sortbyTime, setSortbyTime] = useState('oldest');
  return (
    <OptionList open={open} onOpenChange={setOpen}>
      <OptionList.Trigger>
        <div>
          <div className="hidden md:flex">
            <Toolbar.Button
              content="Sortby"
              variant="basic"
              prefix={ArrowsDownUp}
            />
          </div>

          <div className="flex md:hidden">
            <Toolbar.IconButton variant="basic" icon={ArrowsDownUp} />
          </div>
        </div>
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.RadioGroup
          value={sortbyProperty}
          onValueChange={setSortbyProperty}
        >
          <OptionList.RadioGroupItem
            value="title"
            onSelect={(e) => e.preventDefault()}
          >
            Provider name
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem
            value="updated"
            onSelect={(e) => e.preventDefault()}
          >
            Updated
          </OptionList.RadioGroupItem>
        </OptionList.RadioGroup>
        <OptionList.Separator />
        <OptionList.RadioGroup value={sortbyTime} onValueChange={setSortbyTime}>
          <OptionList.RadioGroupItem
            showIndicator={false}
            value="oldest"
            onSelect={(e) => e.preventDefault()}
          >
            <ArrowUp size={16} />
            Oldest first
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem
            value="newest"
            showIndicator={false}
            onSelect={(e) => e.preventDefault()}
          >
            <ArrowDown size={16} />
            Newest first
          </OptionList.RadioGroupItem>
        </OptionList.RadioGroup>
      </OptionList.Content>
    </OptionList>
  );
};

const ResourceItemExtraOptions = ({ open, setOpen, onEdit, onDelete }) => {
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
        <OptionList.Separator />
        <OptionList.Item className="!text-text-critical" onSelect={onDelete}>
          <Trash size={16} />
          <span>Delete</span>
        </OptionList.Item>
      </OptionList.Content>
    </OptionList>
  );
};

const AddPopUp = ({ addProviderPopup, setAddProviderPopup }) => {
  const api = useAPIClient();
  const reloadPage = useReload();
  const { user } = useOutletContext();

  const { values, errors, handleSubmit, handleChange, isLoading, resetValues } =
    useForm({
      initialValues: {
        displayName: '',
        name: '',
        provider: 'aws',
        accessKey: '',
        accessSecret: '',
      },
      validationSchema: Yup.object({
        displayName: Yup.string().required(),
        name: Yup.string().required(),
        provider: Yup.string().required(),
        accessKey: Yup.string().required(),
        accessSecret: Yup.string().required(),
      }),
      onSubmit: async (val) => {
        try {
          const { errors: e } = await api.createProviderSecret({
            secret: getSecretRef({
              metadata: getMetadata({
                name: val.name,
                annotations: {
                  [keyconstants.displayName]: val.displayName,
                  [keyconstants.provider]: val.provider,
                  [keyconstants.author]: user.name,
                },
              }),
              stringData: {
                accessKey: val.accessKey,
                accessSecret: val.accessSecret,
              },
            }),
          });
          if (e) {
            throw e[0];
          }
          toast.success('provider secret created successfully');
          reloadPage();
          setAddProviderPopup(false);
          resetValues();
        } catch (err) {
          toast.error(err.message);
        }
      },
    });
  return (
    <Popup.PopupRoot show={addProviderPopup} onOpenChange={setAddProviderPopup}>
      <Popup.Header>Add new cloud provider</Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            <TextInput
              label="Name"
              onChange={handleChange('displayName')}
              error={!!errors.displayName}
              message={!!errors.displayName}
              value={values.displayName}
              name="provider-secret-name"
            />
            <IdSelector
              name={values.displayName}
              resType={idTypes.providersecret}
              onChange={(id) => {
                handleChange('name')({ target: { value: id } });
              }}
            />
            <SelectInput.Select
              error={!!errors.provider}
              message={errors.provider}
              value={values.provider}
              label="Provider"
              onChange={(provider) => {
                handleChange('provider')({ target: { value: provider } });
              }}
            >
              <SelectInput.Option value="aws">
                Amazon Web Services
              </SelectInput.Option>
            </SelectInput.Select>
            <PasswordInput
              name="accessKey"
              onChange={handleChange('accessKey')}
              error={!!errors.accessKey}
              message={errors.accessKey}
              value={values.accessKey}
              label="Access Key ID"
            />
            <PasswordInput
              name="accessSecret"
              label="Access Key Secret"
              onChange={handleChange('accessSecret')}
              error={!!errors.accessSecret}
              message={errors.accessSecret}
              value={values.accessSecret}
            />
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button content="Cancel" variant="basic" closable />
          <Popup.Button
            loading={isLoading}
            type="submit"
            content="Add"
            variant="primary"
          />
        </Popup.Footer>
      </form>
    </Popup.PopupRoot>
  );
};
const UpdatePopUp = ({ editProviderPopup, setEditProviderPopup, secret }) => {
  const api = useAPIClient();
  const reloadPage = useReload();
  const { user } = useOutletContext();

  const { values, errors, isLoading, handleChange, handleSubmit, resetValues } =
    useForm({
      initialValues: {
        displayName: parseDisplaynameFromAnn(secret),
        accessSecret: secret?.stringData?.accessSecret || '',
        accessKey: secret?.stringData?.accessKey || '',
      },
      validationSchema: Yup.object({
        displayName: Yup.string().trim().required(),
        accessSecret: Yup.string().trim().required(),
        accessKey: Yup.string().trim().required(),
      }),
      onSubmit: async (val) => {
        try {
          const { errros: e } = await api.updateProviderSecret({
            secret: getSecretRef({
              metadata: getMetadata({
                name: parseName(secret),
                annotations: {
                  [keyconstants.displayName]: val.displayName,
                  [keyconstants.provider]: parseFromAnn(
                    secret,
                    keyconstants.provider
                  ),
                  [keyconstants.author]: user.name,
                },
              }),
              stringData: {
                accessKey: val.accessKey,
                accessSecret: val.accessSecret,
              },
            }),
          });
          if (e) {
            throw e[0];
          }
          toast.success('updated successfully');
          reloadPage();
          resetValues();
          setEditProviderPopup(false);
        } catch (err) {
          toast.error(err.message);
        }
      },
    });
  return (
    <Popup.PopupRoot
      show={editProviderPopup}
      onOpenChange={setEditProviderPopup}
    >
      <Popup.Header>Edit cloud provider</Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            <Chips.Chip
              {...{
                item: { id: parseName(secret) },
                label: parseName(secret),
                prefix: 'Id:',
                disabled: true,
                type: Chips.ChipType.BASIC,
              }}
            />
            <TextInput
              label="Name"
              value={values.displayName}
              error={!!errors.displayName}
              message={errors.displayName}
              onChange={handleChange('displayName')}
            />

            <TextInput
              label="Access Key ID"
              value={values.accessKey}
              error={!!errors.accessKey}
              message={errors.accessKey}
              onChange={handleChange('accessKey')}
              type="password"
            />
            <TextInput
              label="Access Key Secret"
              value={values.accessSecret}
              error={!!errors.accessSecret}
              message={errors.accessSecret}
              onChange={handleChange('accessSecret')}
              type="password"
            />
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button content="Cancel" variant="basic" closable />
          <Popup.Button
            loading={isLoading}
            type="submit"
            content="Update"
            variant="primary"
          />
        </Popup.Footer>
      </form>
    </Popup.PopupRoot>
  );
};

const CloudProvidersIndex = () => {
  const [appliedFilters, setAppliedFilters] = useState(
    dummyData.appliedFilters
  );
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(15);
  const [totalItems, setTotalItems] = useState(100);
  const [viewMode, setViewMode] = useState('list');
  const [addProviderPopup, setAddProviderPopup] = useState(false);

  const { providerSecrets } = useLoaderData();

  useLog(providerSecrets);

  const cloudProviders = providerSecrets?.edges?.map(({ node }) => node) || [];

  return (
    <>
      <SubHeader
        title="Cloud Providers"
        actions={
          cloudProviders.length !== 0 && (
            <Button
              variant="primary"
              content="Create Cloud Provider"
              prefix={PlusFill}
              type="button"
              onClick={() => setAddProviderPopup(true)}
            />
          )
        }
      />
      {cloudProviders.length > 0 && (
        <div className="pt-3xl flex flex-col gap-6xl">
          <div className="flex flex-col">
            <ClusterToolbar viewMode={viewMode} setViewMode={setViewMode} />
            <ClusterFilters
              appliedFilters={appliedFilters}
              setAppliedFilters={setAppliedFilters}
            />
          </div>
          <ResourceList mode={viewMode}>
            {cloudProviders.map((secret) => (
              <ResourceList.ResourceItem
                key={parseUpdationTime(secret) + parseName(secret)}
              >
                <ResourceItem secret={secret} />
              </ResourceList.ResourceItem>
            ))}
          </ResourceList>
          <div className="hidden md:flex">
            <Pagination
              currentPage={currentPage}
              itemsPerPage={itemsPerPage}
              totalItems={totalItems}
            />
          </div>
        </div>
      )}
      {cloudProviders.length === 0 && (
        <div className="pt-3xl">
          <EmptyState
            illustration={
              <svg
                width="226"
                height="227"
                viewBox="0 0 226 227"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <rect y="0.970703" width="226" height="226" fill="#F4F4F5" />
              </svg>
            }
            heading="This is the place where you will oversees the Cloud Provider."
            action={{
              content: 'Create new cloud provider',
              prefix: Plus,
              onClick: () => setAddProviderPopup(true),
            }}
          >
            <p>
              You have the option to include a new Cloud Provider and oversee
              the existing Cloud Provider.
            </p>
          </EmptyState>
        </div>
      )}

      {/* Popup dialog for adding cloud provider */}
      <AddPopUp
        {...{
          addProviderPopup,
          setAddProviderPopup,
        }}
      />
    </>
  );
};

export const loader = async (ctx) => {
  const { data, errors } = await GQLServerHandler(
    ctx.request
  ).listProviderSecrets({
    pagination: getPagination(ctx),
  });

  if (errors) {
    logger.error(errors);
  }

  return {
    providerSecrets: data,
  };
};

export default CloudProvidersIndex;
