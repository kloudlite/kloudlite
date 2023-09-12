import { useParams } from '@remix-run/react';
import { useState } from 'react';
import Popup from '~/components/molecule/popup';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import ConfigResource from '~/console/page-components/config-resource';
import {
  ArrowDown,
  ArrowLeft,
  ArrowUp,
  ArrowsDownUp,
  Search,
  Spinner,
} from '@jengaicons/react';
import { IconButton } from '~/components/atoms/button';
import Toolbar from '~/components/atoms/toolbar';
import OptionList from '~/components/atoms/option-list';
import SecretResource from '~/console/page-components/secret-resource';
import { handleError } from '~/root/lib/utils/common';
import { parseName, parseNodes } from '~/console/server/r-utils/common';
import { IDialog } from '~/console/components/types.d';
import CSComponent from './cs-item';
import { IAppDialogValue } from './route';

const SortbyOptionList = () => {
  const [orderBy, _setOrderBy] = useState('updateTime');
  return (
    <OptionList.Root>
      <OptionList.Trigger>
        <div>
          <div className="hidden md:flex">
            <Toolbar.Button
              content="Sortby"
              variant="basic"
              prefix={<ArrowsDownUp />}
            />
          </div>

          <div className="flex md:hidden">
            <Toolbar.IconButton variant="basic" icon={<ArrowsDownUp />} />
          </div>
        </div>
      </OptionList.Trigger>
      <OptionList.Content>
        <OptionList.RadioGroup>
          <OptionList.RadioGroupItem
            value="metadata.name"
            onSelect={(e) => e.preventDefault()}
          >
            Name
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem
            value="updateTime"
            onSelect={(e) => e.preventDefault()}
          >
            Updated
          </OptionList.RadioGroupItem>
        </OptionList.RadioGroup>
        <OptionList.Separator />
        <OptionList.RadioGroup>
          <OptionList.RadioGroupItem
            showIndicator={false}
            value="ASC"
            onSelect={(e) => e.preventDefault()}
          >
            <ArrowUp size={16} />
            {orderBy === 'updateTime' ? 'Oldest' : 'Ascending'}
          </OptionList.RadioGroupItem>
          <OptionList.RadioGroupItem
            value="DESC"
            showIndicator={false}
            onSelect={(e) => e.preventDefault()}
          >
            <ArrowDown size={16} />
            {orderBy === 'updateTime' ? 'Newest' : 'Descending'}
          </OptionList.RadioGroupItem>
        </OptionList.RadioGroup>
      </OptionList.Content>
    </OptionList.Root>
  );
};

const AppDialog = ({ show, setShow, onSubmit }: IDialog<IAppDialogValue>) => {
  const api = useAPIClient();

  const [isloading, setIsloading] = useState<boolean>(true);
  const { workspace, project, scope } = useParams();

  const [configs, setConfigs] = useState<Array<any>>([]);
  const [showConfig, setShowConfig] = useState<boolean>(false);
  const [selectedConfig, setSelectedConfig] = useState<any>(null);
  const [selectedKey, setSelectedKey] = useState<any>(null);

  const isConfigItemPage = () => {
    return selectedConfig && showConfig;
  };

  const reset = () => {
    setConfigs([]);
    setIsloading(true);
    setSelectedConfig(null);
    setSelectedKey(null);
    setShowConfig(false);
  };

  useDebounce(
    async () => {
      if (!['secret', 'config'].includes(show?.type || '')) {
        return;
      }
      try {
        setIsloading(true);
        let apiCall = api.listConfigs;
        if (show?.type === 'secret') apiCall = api.listSecrets;

        const { data, errors } = await apiCall({
          project: {
            value: project,
            type: 'name',
          },
          scope: {
            value: workspace,
            type: scope === 'workspace' ? 'workspaceName' : 'environmentName',
          },
          //   pagination: getPagination(ctx),
          //   search: getSearch(ctx),
        });
        if (errors) {
          throw errors[0];
        }
        console.log(data);
        setConfigs(parseNodes(data));
      } catch (err) {
        handleError(err);
      } finally {
        setIsloading(false);
      }
    },
    300,
    [show]
  );

  return (
    <Popup.Root
      show={show as any}
      className="!w-[900px]"
      onOpenChange={(e) => {
        if (!e) {
          //   resetValues();
        }
        setShow(e);
      }}
    >
      <Popup.Header showclose={false}>
        <div className="flex flex-row items-center gap-lg">
          {isConfigItemPage() && (
            <IconButton
              size="sm"
              icon={<ArrowLeft />}
              variant="plain"
              onClick={() => {
                setIsloading(true);
                setShowConfig(false);
                setSelectedConfig(null);
                setSelectedKey(null);
                setTimeout(() => {
                  setIsloading(false);
                }, 150);
              }}
            />
          )}
          <div className="flex-1">
            {isConfigItemPage()
              ? parseName(selectedConfig as any)
              : 'Select config'}
          </div>
          <div className="bodyMd text-text-strong font-normal">1/2</div>
        </div>
      </Popup.Header>
      <Popup.Content>
        {!isloading && (
          <div className="flex flex-col gap-3xl">
            <Toolbar.Root>
              <div className="flex-1">
                <Toolbar.TextInput
                  prefixIcon={<Search />}
                  placeholder="Search"
                />
              </div>
              <SortbyOptionList />
            </Toolbar.Root>
            {isConfigItemPage() && (
              <CSComponent
                items={selectedConfig?.data}
                type="config"
                onClick={(val) => {
                  setSelectedKey(val);
                }}
              />
            )}
            {!isloading &&
              !isConfigItemPage() &&
              (show?.type === 'config' ? (
                <ConfigResource
                  items={configs}
                  hasActions={false}
                  onClick={(val) => {
                    setSelectedConfig(val);
                    setIsloading(true);
                    setShowConfig(true);
                    setTimeout(() => {
                      setIsloading(false);
                    }, 150);
                  }}
                  onDelete={() => {}}
                />
              ) : (
                <SecretResource
                  items={configs}
                  hasActions={false}
                  onClick={(val) => {
                    setSelectedConfig(val);
                  }}
                  onDelete={() => {}}
                />
              ))}
          </div>
        )}

        {isloading && (
          <div className="min-h-[100px] flex flex-col items-center justify-center gap-xl">
            <span className="animate-spin">
              <Spinner color="currentColor" weight={2} size={24} />
            </span>
            <span className="text-text-soft bodyMd">Loading</span>
          </div>
        )}
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button
          closable
          content="Cancel"
          variant="basic"
          onClick={() => {
            reset();
          }}
        />
        <Popup.Button
          type="submit"
          content={isConfigItemPage() ? 'Add' : 'Continue'}
          variant="primary"
          disabled={isConfigItemPage() ? !selectedKey : !selectedConfig}
          onClick={() => {
            if (selectedKey && onSubmit) {
              const sK = selectedKey;
              const sC = selectedConfig;
              reset();
              onSubmit({
                refKey: parseName(sC),
                refName: sK,
                type: 'config',
              });
            }
          }}
        />
      </Popup.Footer>
    </Popup.Root>
  );
};

export default AppDialog;
