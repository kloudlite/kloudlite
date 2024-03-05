/* eslint-disable no-nested-ternary */
import { Spinner } from '@jengaicons/react';
import { useParams } from '@remix-run/react';
import { useState } from 'react';
import { IconButton } from '~/components/atoms/button';
import Popup from '~/components/molecule/popup';
import MultiStep, { useMultiStep } from '~/console/components/multi-step';
import NoResultsFound from '~/console/components/no-results-found';
import { IDialog } from '~/console/components/types.d';
import ConfigResource from '~/console/page-components/config-resource';
import SecretResources from '~/console/page-components/secret-resource';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { IConfigs } from '~/console/server/gql/queries/config-queries';
import { ISecrets } from '~/console/server/gql/queries/secret-queries';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
} from '~/console/server/r-utils/common';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import { handleError } from '~/root/lib/utils/common';
import { ArrowLeft } from '~/console/components/icons';
import { IAppDialogValue } from './app-environment';
import CSComponent, { ICSComponent } from './cs-item';

const AppDialog = ({
  show,
  setShow,
  onSubmit,
}: IDialog<null, IAppDialogValue>) => {
  const api = useConsoleApi();
  const {
    currentStep,
    onNext,
    onPrevious,
    reset: resetStep,
  } = useMultiStep({
    defaultStep: 1,
    totalSteps: 2,
  });

  const [isloading, setIsloading] = useState<boolean>(true);
  const { environment, project } = useParams();

  const [configs, setConfigs] = useState<
    ExtractNodeType<ISecrets>[] | ExtractNodeType<IConfigs>[]
  >([]);
  const [selectedConfig, setSelectedConfig] = useState<
    (ExtractNodeType<ISecrets> & ExtractNodeType<IConfigs>) | null
  >(null);
  const [selectedKey, setSelectedKey] = useState<string>('');

  const reset = () => {
    setConfigs([]);
    setIsloading(true);
    setSelectedConfig(null);
    setSelectedKey('');
    resetStep();
  };

  const showLoading = () => {
    setIsloading(true);
    setTimeout(() => {
      setIsloading(false);
    }, 150);
  };

  useDebounce(
    async () => {
      if (
        !['secret', 'config'].includes(show?.type || '') ||
        !(environment && project)
      ) {
        return;
      }
      try {
        setIsloading(true);
        let apiCall = api.listConfigs;
        if (show?.type === 'secret') apiCall = api.listSecrets;

        const { data, errors } = await apiCall({
          projectName: project,
          envName: environment,
          //   pagination: getPagination(ctx),
          //   search: getSearch(ctx),
        });
        if (errors) {
          throw errors[0];
        }
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
          {currentStep === 2 && (
            <IconButton
              size="sm"
              icon={<ArrowLeft />}
              variant="plain"
              onClick={() => {
                setSelectedConfig(null);
                setSelectedKey('');
                showLoading();
                onPrevious();
              }}
            />
          )}
          <div className="flex-1">
            {currentStep === 2
              ? selectedConfig?.displayName
              : show?.type === 'config'
              ? 'Select config'
              : 'Select secret'}
          </div>
          <div className="bodyMd text-text-strong font-normal">
            {currentStep}/2
          </div>
        </div>
      </Popup.Header>
      <Popup.Content>
        {!isloading && (
          <div className="min-h-[40vh]">
            <MultiStep.Root currentStep={currentStep}>
              <MultiStep.Step step={1}>
                {configs.length > 0 ? (
                  show?.type === 'config' ? (
                    <ConfigResource
                      items={configs}
                      hasActions={false}
                      onClick={(val) => {
                        setSelectedConfig(val);
                        showLoading();
                        onNext();
                      }}
                    />
                  ) : (
                    <SecretResources
                      items={configs}
                      hasActions={false}
                      onClick={(val) => {
                        setSelectedConfig(val);
                        showLoading();
                        onNext();
                      }}
                    />
                  )
                ) : (
                  <NoResultsFound
                    title={
                      show?.type === 'config'
                        ? 'No configs are added.'
                        : 'No secrets are added.'
                    }
                    subtitle={
                      show?.type === 'config'
                        ? 'Please add configs in Configs and Secrets.'
                        : 'Please add secrets in Configs and Secrets.'
                    }
                    shadow={false}
                    border={false}
                  />
                )}
              </MultiStep.Step>
              <MultiStep.Step step={2}>
                {Object.keys(
                  (show?.type === 'config'
                    ? selectedConfig?.data
                    : selectedConfig?.stringData) || {}
                ).length !== 0 ? (
                  <CSComponent
                    items={
                      show?.type === 'config'
                        ? selectedConfig?.data
                        : selectedConfig?.stringData
                    }
                    type={show?.type as ICSComponent['type']}
                    onClick={(val) => {
                      setSelectedKey(val);
                    }}
                  />
                ) : (
                  <NoResultsFound
                    title="Data not available."
                    subtitle="Please add in Configs and Secrets."
                    shadow={false}
                    border={false}
                  />
                )}
              </MultiStep.Step>
            </MultiStep.Root>
          </div>
        )}

        {isloading && (
          <div className="min-h-[40vh] flex flex-col items-center justify-center gap-xl">
            <span className="animate-spin">
              <Spinner color="currentColor" weight={2} size={24} />
            </span>
            <span className="text-text-soft bodyMd">Loading</span>
          </div>
        )}
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button
          closable={currentStep === 1}
          content={currentStep === 1 ? 'Cancel' : 'Back'}
          variant="basic"
          onClick={() => {
            if (currentStep === 1) {
              reset();
            } else {
              showLoading();
              onPrevious();
              setSelectedConfig(null);
            }
          }}
        />
        <Popup.Button
          type="submit"
          content={currentStep === 2 ? 'Add' : 'Continue'}
          variant="primary"
          disabled={currentStep === 1 ? !selectedConfig : !selectedKey}
          onClick={() => {
            if (currentStep === 2 && selectedKey && onSubmit) {
              const sK = selectedKey;
              const sC = selectedConfig;
              reset();
              onSubmit({
                refKey: sK,
                refName: parseName(sC),
                type: show?.type as ICSComponent['type'],
              });
            }
          }}
        />
      </Popup.Footer>
    </Popup.Root>
  );
};

export default AppDialog;
