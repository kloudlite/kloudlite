import { toast } from '~/components/molecule/toast';
import logger from '../client/helpers/log';

export const handleError = (e: unknown): void => {
  const err = e as Error;
  logger.error(e);
  toast.error(err.message);
};

export const parseError = (e: unknown): Error => {
  return e as Error;
};
