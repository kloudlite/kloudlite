import { useOutletContext, useSearchParams } from '@remix-run/react';
import HighlightJsLog from '~/console/components/logger';
import { NumberInput } from '~/components/atoms/input';
import { useState } from 'react';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { Button } from '~/components/atoms/button';
import { useQueryParameters } from '~/root/lib/client/hooks/use-search';
import { parseName } from '~/console/server/r-utils/common';
import { IAppContext } from '../_.$account.$cluster.$project.$scope.$workspace.app.$app/route';

const ItemList = () => {
  const { app } = useOutletContext<IAppContext>();
  const [sp] = useSearchParams();

  const [url] = useState(
    `wss://observability.dev.kloudlite.io/observability/logs/app?resource_name=${parseName(
      app
    )}&resource_namespace=${app.metadata!.namespace}&start_time=${
      sp.get('start') || 1690273382
    }&end_time=${sp.get('end') || 1690532560}`
  );

  const { setQueryParameters } = useQueryParameters();

  const { values, handleChange, handleSubmit } = useForm({
    initialValues: {
      start: sp.get('start') || '1690273382',
      end: sp.get('end') || '1690532560',
    },
    validationSchema: Yup.object({}),
    onSubmit: (val) => {
      // @ts-ignore
      setQueryParameters(val);
    },
  });

  return (
    <div className="p-lg flex flex-col gap-xl">
      <div>Logs Url: {url}</div>
      <HighlightJsLog
        dark
        websocket
        height="60vh"
        width="100%"
        url={url}
        selectableLines
      />
      <form onSubmit={handleSubmit} className="flex flex-col gap-xl">
        <NumberInput
          label="start data timestamp"
          value={values.start}
          onChange={handleChange('start')}
          placeholder="start"
        />
        <NumberInput
          label="end date timestamp"
          placeholder="end"
          value={values.end}
          onChange={handleChange('end')}
        />
        <div className="flex gap-xl">
          <Button type="submit" content="update search params" />
          <Button
            onClick={() => {
              window.location.reload();
            }}
            content="reload"
          />
        </div>
      </form>
    </div>
  );
};

export default ItemList;
