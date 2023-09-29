import { toast } from '~/components/molecule/toast';
import { titleCase } from '~/components/utils';
import logger from '../client/helpers/log';

export const handleError = (e: unknown): void => {
  const err = e as Error;
  logger.error(e);
  toast.error(err.message);
};

export const parseError = (e: unknown): Error => {
  return e as Error;
};

export const truncate = (str: string, length: number) => {
  return str.length > length ? `${str.substring(0, length)}...` : str;
};

export const Truncate = ({
  children: str,
  length,
}: {
  children: string;
  length: number;
}) => {
  const newStr = str.length > length ? `${str.substring(0, length)}...` : str;
  return <span title={str}>{newStr}</span>;
};
