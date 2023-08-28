import { useParams } from '@remix-run/react';
import { useState } from 'react';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { dummyData } from '~/console/dummy/data';
import { parseName, parseNodes } from '~/console/server/r-urils/common';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import useMatches, {
  useDataFromMatches,
} from '~/root/lib/client/hooks/use-custom-matches';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import ConfigResource from '~/console/page-components/config-resource';
import { ArrowLeft, Spinner } from '@jengaicons/react';
import { AnimatePresence, LayoutGroup, motion } from 'framer-motion';
import { IconButton } from '~/components/atoms/button';
import ResourcesConfig from './resource-config';

const Main = ({ show, setShow }) => {
  const api = useAPIClient();
  const [isloading, setIsloading] = useState(true);
  const { workspace, project, scope } = useParams();
  const [configs, setConfigs] = useState([]);
  const [showConfig, setShowConfig] = useState(null);
  const [selectedConfig, setSelectedConfig] = useState(null);
  const [selectedKey, setSelectedKey] = useState(null);

  useDebounce(
    async () => {
      try {
        setIsloading(true);
        const { data, errors } = await api.listConfigs({
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
        setConfigs(parseNodes(data));
      } catch (err) {
        toast.error(err.message);
      } finally {
        setIsloading(false);
      }
    },
    300,
    []
  );

  return (
    <Popup.Root
      show={show}
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
          {showConfig && (
            <IconButton
              icon={ArrowLeft}
              variant="plain"
              onClick={() => {
                setShowConfig(false);
                setSelectedConfig(null);
                setSelectedKey(null);
              }}
            />
          )}
          <div className="flex-1">
            {showConfig ? parseName(selectedConfig) : 'Select config'}
          </div>
          <div className="bodyMd text-text-strong font-normal">1/2</div>
        </div>
      </Popup.Header>
      <Popup.Content>
        <>
          <AnimatePresence>
            {!isloading && (
              <motion.div
                key={selectedConfig && showConfig ? 1 : 0}
                initial={{
                  opacity: 0,
                }}
                animate={{
                  opacity: 1,
                }}
                exit={{
                  opacity: 0,
                }}
                transition={{
                  ease: 'anticipate',
                }}
              >
                {selectedConfig && showConfig ? (
                  <ResourcesConfig
                    items={selectedConfig?.data}
                    onClick={(val) => {
                      setSelectedKey(val);
                    }}
                  />
                ) : (
                  <ConfigResource
                    items={configs}
                    hasActions={false}
                    onClick={(val) => {
                      setSelectedConfig(val);
                    }}
                  />
                )}
              </motion.div>
            )}
          </AnimatePresence>

          {isloading && (
            <div className="min-h-[100px] flex flex-col items-center justify-center gap-xl">
              <span className="animate-spin">
                <Spinner color="currentColor" weight={2} size={24} />
              </span>
              <span className="text-text-soft bodyMd">Loading</span>
            </div>
          )}
        </>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button closable content="Cancel" variant="basic" />
        <Popup.Button
          type="submit"
          content={showConfig ? 'Add' : 'Continue'}
          variant="primary"
          disabled={showConfig ? !selectedKey : !selectedConfig}
          onClick={() => {
            if (selectedConfig) {
              setShowConfig(true);
            }
          }}
        />
      </Popup.Footer>
    </Popup.Root>
  );
};

const HandleConfig = ({ show, setShow }) => {
  if (show) {
    return <Main show={show} setShow={setShow} />;
  }
  return null;
};

export default HandleConfig;
