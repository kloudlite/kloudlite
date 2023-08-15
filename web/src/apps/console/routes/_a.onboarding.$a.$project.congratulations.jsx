import {
  Link,
  useOutletContext,
  useParams,
  useLoaderData,
} from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import logger from '~/root/lib/client/helpers/log';
import { GQLServerHandler } from '../server/gql/saved-queries';
import { parseDisplaynameFromAnn, parseName } from '../server/r-urils/common';

const Congratulations = () => {
  // @ts-ignore
  const { account } = useOutletContext();
  const { a: accountName } = useParams();
  const { project } = useLoaderData();
  return (
    <div className="px-11xl pt-8xl pb-10xl flex items-center justify-center h-full">
      <div className="flex flex-col gap-6xl w-[450px] text-center">
        <div className="bg-surface-basic-active h-[200px]" />
        <div className="flex flex-col gap-3xl">
          <div className="heading4xl text-text-default">Congratulations 🚀</div>
          <div className="bodyLg text-text-soft block">
            You’ve successfully create your organization{' '}
            <span className="headingMd">
              {account.displayName || account.name}
            </span>{' '}
            and deployed first project{' '}
            <span className="headingMd">
              {parseDisplaynameFromAnn(project) || parseName(project)}
            </span>
            .
          </div>
          <Button
            LinkComponent={Link}
            href={`/${accountName}/projects`}
            content="Continue to dashboard"
            variant="basic"
            block
          />
        </div>
      </div>
    </div>
  );
};

export const loader = async (ctx) => {
  const { project } = ctx.params;
  try {
    const { data, errors } = await GQLServerHandler(ctx.request).getProject({
      name: project,
    });
    if (errors) {
      throw errors[0];
    }
    return {
      project: data,
    };
  } catch (err) {
    logger.error(err);
  }

  return {};
};

export default Congratulations;
