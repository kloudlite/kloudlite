import { useParams } from '@remix-run/react';
import { useState } from 'react';
import Popup from '~/components/molecule/popup';
import { toast } from '~/components/molecule/toast';
import { parseName, parseNodes } from '~/console/server/r-urils/common';
import { useAPIClient } from '~/root/lib/client/hooks/api-provider';
import useDebounce from '~/root/lib/client/hooks/use-debounce';
import ConfigResource from '~/console/page-components/config-resource';
import { ArrowLeft, Spinner } from '@jengaicons/react';
import { AnimatePresence, motion } from 'framer-motion';
import { IconButton } from '~/components/atoms/button';
import { handleError } from '~/root/lib/types/common';
import ConfigItem from './config-item';
import ResourcesConfig from './resource-config';

const AnimatePage = ({ children, visible }) => {
  return (
    <AnimatePresence>
      {visible && (
        <motion.div
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
          {children}
        </motion.div>
      )}
    </AnimatePresence>
  );
};

const Main = ({ show, setShow, onSubmit = (_) => _ }) => {
  const api = useAPIClient();

  const [isloading, setIsloading] = useState(true);
  const { workspace, project, scope } = useParams();

  const [configs, setConfigs] = useState([]);
  const [showConfig, setShowConfig] = useState(null);
  const [selectedConfig, setSelectedConfig] = useState(null);
  const [selectedKey, setSelectedKey] = useState(null);

  const isConfigItemPage = () => {
    return selectedConfig && showConfig;
  };

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
          {isConfigItemPage() && (
            <IconButton
              size="sm"
              icon={<ArrowLeft />}
              variant="plain"
              onClick={() => {
                setShowConfig(false);
                setSelectedConfig(null);
                setSelectedKey(null);
              }}
            />
          )}
          <div className="flex-1">
            {isConfigItemPage() ? parseName(selectedConfig) : 'Select config'}
          </div>
          <div className="bodyMd text-text-strong font-normal">1/2</div>
        </div>
      </Popup.Header>
      <Popup.Content>
        <>
          <AnimatePresence mode="wait">
            <motion.div
              key={isConfigItemPage() ? 'configitempage' : 'configpage'}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
            >
              {isConfigItemPage() && (
                <ConfigItem
                  items={selectedConfig?.data}
                  onClick={(val) => {
                    setSelectedKey(val);
                  }}
                />
              )}
              {!isloading && !isConfigItemPage() && (
                <ConfigResource
                  items={configs}
                  hasActions={false}
                  onClick={(val) => {
                    setSelectedConfig(val);
                  }}
                />
              )}
            </motion.div>
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
          content={isConfigItemPage() ? 'Add' : 'Continue'}
          variant="primary"
          disabled={isConfigItemPage() ? !selectedKey : !selectedConfig}
          onClick={() => {
            if (selectedConfig) {
              setShowConfig(true);
            }
            if (selectedKey) {
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

const AppDialog = ({ show, setShow, onSubmit }) => {
  if (show) {
    return <Main show={show} setShow={setShow} onSubmit={onSubmit} />;
  }
  return null;
};

export default AppDialog;
