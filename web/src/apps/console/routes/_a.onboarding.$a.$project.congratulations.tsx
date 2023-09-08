import {
  Link,
  useOutletContext,
  useParams,
  useLoaderData,
} from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import { IRemixCtx } from '~/root/lib/types/common';
import { defer } from '@remix-run/node';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { LoadingComp, pWrapper } from '../components/loading-component';
import { parseName } from '../server/r-utils/common';

export const loader = async (ctx: IRemixCtx) => {
  const { project } = ctx.params;

  const promise = pWrapper(async () => {
    const { data, errors } = await GQLServerHandler(ctx.request).getProject({
      name: project,
    });
    if (errors) {
      throw errors[0];
    }
    return {
      project: data,
    };
  });

  return defer({
    promise,
  });
};

const Congratulations = () => {
  // @ts-ignore
  const { account } = useOutletContext();
  const { a: accountName } = useParams();
  const { promise } = useLoaderData<typeof loader>();
  return (
    <LoadingComp data={promise}>
      {({ project }) => {
        return (
          <div className="px-11xl pt-8xl pb-10xl flex items-center justify-center h-full">
            <div className="flex flex-col gap-6xl w-[450px] text-center">
              <div className="bg-surface-basic-active h-[200px]" />
              <div className="flex flex-col gap-3xl">
                <div className="heading4xl text-text-default">
                  Congratulations 🚀
                </div>
                <div className="bodyLg text-text-soft block">
                  You’ve successfully create your organization{' '}
                  <span className="headingMd">
                    {account.displayName || account.name}
                  </span>{' '}
                  and deployed first project{' '}
                  <span className="headingMd">
                    {project.displayName || parseName(project)}
                  </span>
                  .
                </div>
                <Button
                  LinkComponent={Link}
                  to={`/${accountName}/projects`}
                  content="Continue to dashboard"
                  variant="basic"
                  block
                />
              </div>
            </div>
          </div>
        );
      }}
    </LoadingComp>
  );
};

export default Congratulations;
