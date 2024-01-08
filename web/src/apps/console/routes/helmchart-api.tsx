import axios from 'axios';
import getQueries from '~/root/lib/server/helpers/get-queries';
import { IExtRemixCtx } from '~/root/lib/types/common';

export const loader = async (ctx: IExtRemixCtx) => {
  const { url } = getQueries(ctx);
  const axiosInstance = axios.create({
    maxRedirects: 1000,
  });
  const res = await axiosInstance.get(`${url}/index.yaml`);
  return res.data;
};
