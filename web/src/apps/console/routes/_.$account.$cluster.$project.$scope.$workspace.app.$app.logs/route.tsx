import { useOutletContext } from '@remix-run/react';
import HighlightJsLog from '~/console/components/logger';
import { IAppContext } from '../_.$account.$cluster.$project.$scope.$workspace.app.$app/route';

const ItemList = () => {
  const { app } = useOutletContext<IAppContext>();

  return (
    <div className="p-lg">
      <HighlightJsLog
        dark
        websocket
        height="80vh"
        width="100%"
        url={`wss://observability.dev.kloudlite.io/observability/logs/app?resource_name=${app.metadata.name}&resource_namespace=${app.metadata.namespace}&start_time=1690273382&end_time=1690532560`}
      />
    </div>
  );
};

export default ItemList;
