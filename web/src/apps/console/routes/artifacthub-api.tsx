import axios from 'axios';
import { artifactEnv } from '~/root/lib/configs/envs';
import getQueries from '~/root/lib/server/helpers/get-queries';
import { IExtRemixCtx } from '~/root/lib/types/common';

export const loader = async (ctx: IExtRemixCtx) => {
  const { offest, limit, kind, ts_query_web } = getQueries(ctx);
  const res = await axios.get(`https://artifacthub.io/api/v1/packages/search`, {
    params: {
      offest,
      limit,
      kind,
      ts_query_web,
    },
    headers: {
      'X-API-KEY-ID': artifactEnv.artifact_hub_key_id,
      'X-API-KEY-SECRET': artifactEnv.artifact_hub_key_secret,
    },
  });

  return res.data;
};
