import ExtendedFilledTab from '~/iotconsole/components/extended-filled-tab';
import { AnimatePresence, motion } from 'framer-motion';
import {
  createAppEnvPage,
  useAppState,
} from '~/iotconsole/page-components/app-states';
import AppWrapper from '~/iotconsole/page-components/app/app-wrapper';
import { EnvironmentVariables } from '~/iotconsole/routes/_main+/$account+/$project+/deviceblueprint+/$deviceblueprint+/new-app/app-environment-variables';
import { ConfigMounts } from '~/iotconsole/routes/_main+/$account+/$project+/deviceblueprint+/$deviceblueprint+/new-app/app-environment-mounts';

export interface IAppDialogValue {
  refKey: string;
  refName: string;
  type: 'config' | 'secret';
}

const SettingEnvironment = () => {
  const { envPage, setEnvPage } = useAppState();
  const items: {
    label: string;
    value: createAppEnvPage;
  }[] = [
    {
      label: `Environment variables`,
      value: 'environment_variables',
    },
    {
      label: 'Config files',
      value: 'config_mounts',
    },
  ];

  return (
    <AppWrapper title="Environment">
      <ExtendedFilledTab value={envPage} onChange={setEnvPage} items={items} />
      <AnimatePresence mode="wait">
        <motion.div
          key={envPage || 'empty'}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.15 }}
          className="flex flex-col gap-6xl w-full"
        >
          {envPage === 'environment_variables' && <EnvironmentVariables />}
          {envPage === 'config_mounts' && <ConfigMounts />}
        </motion.div>
      </AnimatePresence>
    </AppWrapper>
  );
};

export default SettingEnvironment;
