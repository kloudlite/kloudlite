import { ConsoleApiType } from '~/console/server/gql/saved-queries';
import { handleError } from '~/root/lib/utils/common';
import { BuildIn, RepositoryIn } from '~/root/src/generated/gql/server';

const createRepo = async ({
  api,
  repository,
}: {
  api: ConsoleApiType;
  repository: RepositoryIn;
}) => {
  try {
    const { errors } = await api.createRepo({
      repository,
    });

    if (errors) {
      throw errors[0];
    }
    return repository.name;
  } catch (err) {
    handleError(err);
    return null;
  }
};

const createBuild = async ({
  api,
  build,
}: {
  api: ConsoleApiType;
  build: BuildIn;
}) => {
  try {
    const { errors, data } = await api.createBuild({
      build,
    });

    if (errors) {
      throw errors[0];
    }
    return data.id;
  } catch (err) {
    handleError(err);
    return null;
  }
};

const updateBuild = async ({
  api,
  build,
  buildId,
}: {
  api: ConsoleApiType;
  build: BuildIn;
  buildId: string;
}) => {
  try {
    const { errors, data } = await api.updateBuild({
      build,
      crUpdateBuildId: buildId,
    });

    if (errors) {
      throw errors[0];
    }
    return data.id;
  } catch (err) {
    handleError(err);
    return null;
  }
};

const triggerBuild = async ({
  api,
  buildId,
}: {
  api: ConsoleApiType;
  buildId: string;
}) => {
  try {
    const { errors } = await api.triggerBuild({
      crTriggerBuildId: buildId,
    });

    if (errors) {
      throw errors[0];
    }
    return buildId;
  } catch (err) {
    handleError(err);
    return null;
  }
};

const appFun = {
  createRepo,
  createBuild,
  updateBuild,
  triggerBuild,
};

export default appFun;
