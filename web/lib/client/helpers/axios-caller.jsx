import { useEffect, useState } from 'react';
import { toast } from '~/components/molecule/toast';
import logger from './log';

const df = async () => {};

export const useCall = () => {
  const [isOperating, setIsOperating] = useState(false);
  const [caller, setCr] = useState(df);
  const [toCall, setCall] = useState(false);

  const call = () => {
    setCall(true);
  };

  const setCaller = (fn) => {
    setCr(fn);
    setCall(true);
  };

  useEffect(() => {
    // @ts-ignore
    if (caller === df || !toCall) {
      return;
    }
    (async () => {
      if (isOperating) {
        toast.warning('processing, please wait...');
        return;
      }
      try {
        // @ts-ignore
        await caller();
        setIsOperating(true);
      } catch (err) {
        toast.error(err.message);
        logger.error(err);
      } finally {
        setIsOperating(false);
      }
    })();
  }, [toCall]);

  return [isOperating, setCaller, call];
};
