import { useParams } from '@remix-run/react';
import { useState } from 'react';
import Popup from '~/components/molecule/popup';
import { parseName, parseNodes } from '~/console/server/r-urils/common';
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
import ConfigItem from './config-item';
import { IValue } from './app-environment';

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

interface IShowBase {
  type: string;
  data: { [key: string]: any } | null;
}

export type IShow = IShowBase | null | boolean;

export interface IDialog<T> {
  show: IShow;
  setShow: React.Dispatch<React.SetStateAction<IShow>>;
  onSubmit?: (data: T) => void;
}

const AppDialog = ({ show, setShow, onSubmit }: IDialog<IValue>) => {
  const api = useAPIClient();

  const [isloading, setIsloading] = useState(true);
  const { workspace, project, scope } = useParams();

  const [configs, setConfigs] = useState<Array<any>>([]);
  const [showConfig, setShowConfig] = useState<boolean>(false);
  const [selectedConfig, setSelectedConfig] = useState<any>(null);
  const [selectedKey, setSelectedKey] = useState<any>(null);

  const isConfigItemPage = () => {
    return selectedConfig && showConfig;
  };

  useDebounce(
    async () => {
      try {
        setIsloading(true);
        let apiCall = api.listConfigs;
        if ((show as IShowBase)?.type === 'secret') apiCall = api.listSecrets;

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
    []
  );

  return (
    <Popup.Root
      show={show as any}
      className="!w-[900px]"
      onOpenChange={(e) => {
        if (!e) {
          //   resetValues();
        }

        setShow(e as IShow);
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
              <ConfigItem
                items={selectedConfig?.data}
                onClick={(val) => {
                  setSelectedKey(val);
                }}
              />
            )}
            {!isloading &&
              !isConfigItemPage() &&
              ((show as IShowBase)?.type === 'config' ? (
                <ConfigResource
                  items={configs}
                  hasActions={false}
                  onClick={(val) => {
                    setSelectedConfig(val);
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
        <Popup.Button closable content="Cancel" variant="basic" />
        <Popup.Button
          type="submit"
          content={isConfigItemPage() ? 'Add' : 'Continue'}
          variant="primary"
          disabled={isConfigItemPage() ? !selectedKey : !selectedConfig}
          onClick={() => {
            if (selectedConfig) {
              setIsloading(true);
              setShowConfig(true);
              setTimeout(() => {
                setIsloading(false);
              }, 150);
            }
            if (selectedKey && onSubmit) {
              onSubmit({
                variable: parseName(selectedConfig),
                key: selectedKey,
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
