import { Plus, PlusFill } from '@jengaicons/react';
import { defer } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button.jsx';
import { LoadingComp, pWrapper } from '~/console/components/loading-component';
import Wrapper from '~/console/components/wrapper';
import { GQLServerHandler } from '~/console/server/gql/saved-queries';
import { parseNodes } from '~/console/server/r-utils/common';
import { IRemixCtx } from '~/root/lib/types/common';
import Tools from './tools';
import HelmChartResources from './helm-chart-resources';
import HandleHelmChart from './handle-helm-chart';

export const loader = (ctx: IRemixCtx) => {
  const { cluster } = ctx.params;
  const promise = pWrapper(async () => {
    const { data: mData, errors: mErrors } = await GQLServerHandler(
      ctx.request
    ).listHelmChart({
      clusterName: cluster,
    });

    if (mErrors) {
      throw mErrors[0];
    }

    return { helmChartData: mData };
  });
  return defer({ promise });
};

const HelmCharts = () => {
  const [visible, setVisible] = useState(false);
  const { promise } = useLoaderData<typeof loader>();

  return (
    <>
      <LoadingComp data={promise}>
        {({ helmChartData }) => {
          const helmCharts = parseNodes(helmChartData);

          return (
            <Wrapper
              secondaryHeader={{
                title: 'Helm charts',
                action: helmCharts.length > 0 && (
                  <Button
                    variant="primary"
                    content="Create helm chart"
                    prefix={<PlusFill />}
                    onClick={() => {
                      setVisible(true);
                    }}
                  />
                ),
              }}
              empty={{
                is: helmCharts.length === 0,
                title: 'This is where you’ll manage your Helm charts.',
                content: (
                  <p>
                    You can create a new backing service and manage the listed
                    backing service.
                  </p>
                ),
                action: {
                  content: 'Create helm chart',
                  prefix: <Plus />,
                  onClick: () => {
                    setVisible(true);
                  },
                },
              }}
              tools={<Tools />}
            >
              <HelmChartResources items={helmCharts} />
            </Wrapper>
          );
        }}
      </LoadingComp>
      <HandleHelmChart
        {...{
          isUpdate: false,
          visible,
          setVisible,
        }}
      />
    </>
  );
};

export default HelmCharts;
