import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import { AnimatePresence, motion } from 'framer-motion';
import { useState } from 'react';
import { EnvironmentVariables } from './app-environment-variables';
import { ConfigMounts } from './app-environment-mounts';
import { FadeIn } from '../_.$account.$cluster.$project.$scope.$workspace.new-app/util';
import { createAppEnvPage } from '../_.$account.$cluster.$project.$scope.$workspace.new-app/states';

export interface IAppDialogValue {
  refKey: string;
  refName: string;
  type: 'config' | 'secret';
}

const SettingEnvironment = () => {
  // const { envPage, setEnvPage, setPage, markPageAsCompleted } = useAppState();
  const [envPage, setEnvPage] = useState('environment_variables');
  const items: {
    label: string;
    value: createAppEnvPage;
  }[] = [
    {
      label: `Environment variables`,
      value: 'environment_variables',
    },
    {
      label: 'Config mount',
      value: 'config_mounts',
    },
  ];

  return (
    <FadeIn
      onSubmit={(e) => {
        e.preventDefault();
      }}
    >
      <div className="flex flex-col gap-xl ">
        <div className="headingXl text-text-default">Environment</div>
        <ExtendedFilledTab
          value={envPage}
          onChange={setEnvPage}
          items={items}
        />
      </div>
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
    </FadeIn>
  );
};

export default SettingEnvironment;
