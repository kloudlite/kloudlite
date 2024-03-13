import { toast } from '~/components/molecule/toast';
import logger from '../client/helpers/log';

export const handleError = (
  e: unknown
): {
  error?: {
    message: string;
  };
} => {
  if (typeof window === 'undefined') {
    const a = e as Error;
    return {
      error: {
        message: a.message,
      },
    };
  }

  const err = e as Error;
  toast.error(err.message);
  logger.error(e);

  return {};
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
  const newStr = str?.length > length ? `${str.substring(0, length)}...` : str;
  return <span title={str}>{newStr}</span>;
};

export function sleep(time: number) {
  return new Promise((resolve) => {
    setTimeout(resolve, time);
  });
}

export const anyUndefined: any = undefined;
