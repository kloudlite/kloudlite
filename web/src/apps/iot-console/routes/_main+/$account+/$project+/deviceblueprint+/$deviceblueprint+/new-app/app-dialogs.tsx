/* eslint-disable no-nested-ternary */
import { useParams } from '@remix-run/react';
import { useState } from 'react';
import { IconButton } from '~/components/atoms/button';
import Popup from '~/components/molecule/popup';
import MultiStep, { useMultiStep } from '~/iotconsole/components/multi-step';
import NoResultsFound from '~/iotconsole/components/no-results-found';
import { IDialog } from '~/iotconsole/components/types.d';
import { useIotConsoleApi } from '~/iotconsole/server/gql/api-provider';
import {
  ExtractNodeType,
  parseName,
  parseNodes,
} from '~/iotconsole/server/r-utils/common';
import useDebounce from '~/lib/client/hooks/use-debounce';
import { handleError } from '~/lib/utils/common';
import { ArrowLeft, Spinner } from '~/iotconsole/components/icons';
import ConfigResourcesV2 from '~/iotconsole/page-components/config-resource-v2';
import SecretResourcesV2 from '~/iotconsole/page-components/secret-resource-v2';
import { ISecrets } from '~/iotconsole/server/gql/queries/iot-secret-queries';
import { IConfigs } from '~/iotconsole/server/gql/queries/iot-config-queries';
import { IAppDialogValue } from './app-environment';
import CSComponent, { ICSComponent } from './cs-item';

const AppDialog = ({
  show,
  setShow,
  onSubmit,
}: IDialog<null, IAppDialogValue>) => {
  const api = useIotConsoleApi();
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

  const [configs, setConfigs] = useState<ExtractNodeType<IConfigs>[]>([]);
  const [secrets, setSecrets] = useState<ExtractNodeType<ISecrets>[]>([]);
  const [mres, setMres] = useState<ExtractNodeType<ISecrets>[]>([]);
  const [selectedConfig, setSelectedConfig] = useState<
    (ExtractNodeType<ISecrets> & ExtractNodeType<IConfigs>) | null
  >(null);
  const [selectedKey, setSelectedKey] = useState<string>('');

  const reset = () => {
    setConfigs([]);
    setSecrets([]);
    setMres([]);
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
        !['secret', 'config', 'mres'].includes(show?.type || '') ||
        !(environment && project)
      ) {
        return;
      }
      try {
        setIsloading(true);

        if (show?.type === 'config') {
          const { data, errors } = await api.listConfigs({
            projectName: project,
            envName: environment,
          });
          if (errors) {
            throw errors[0];
          }
          setConfigs(parseNodes(data));
        } else {
          const { data, errors } = await api.listSecrets({
            projectName: project,
            envName: environment,
          });
          if (errors) {
            throw errors[0];
          }
          if (show?.type === 'secret') {
            setSecrets(parseNodes(data).filter((s) => s.isReadyOnly === false));
          } else {
            setMres(parseNodes(data).filter((m) => m.isReadyOnly === true));
          }
        }
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
              : show?.type === 'secret'
              ? 'Select secret'
              : 'Select managed resource'}
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
                {configs.length > 0 || secrets.length > 0 || mres.length > 0 ? (
                  show?.type === 'config' ? (
                    <ConfigResourcesV2
                      items={configs}
                      hasActions={false}
                      onClick={(val) => {
                        setSelectedConfig({ ...val, isReadyOnly: false });
                        showLoading();
                        onNext();
                      }}
                    />
                  ) : (
                    <SecretResourcesV2
                      items={show?.type === 'secret' ? secrets : mres}
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
                    title={(() => {
                      switch (show?.type) {
                        case 'config':
                          return 'Configs not available.';
                        case 'secret':
                          return 'Secrets not available.';
                        case 'mres':
                          return 'Managed Resources not available.';
                        default:
                          return 'Data not available.';
                      }
                    })()}
                    subtitle={(() => {
                      switch (show?.type) {
                        case 'config':
                          return 'Please add in Configs.';
                        case 'secret':
                          return 'Please add in Secrets.';
                        case 'mres':
                          return 'Please add in Managed Resources.';
                        default:
                          return 'Please add in Configs and Secrets.';
                      }
                    })()}
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
                // type: show?.type as ICSComponent['type'],
                type: (() => {
                  switch (show?.type) {
                    case 'config':
                      return 'config';
                    case 'secret':
                    case 'mres':
                      return 'secret';
                    default:
                      return 'config';
                  }
                })() as ICSComponent['type'],
              });
            }
          }}
        />
      </Popup.Footer>
    </Popup.Root>
  );
};

export default AppDialog;
