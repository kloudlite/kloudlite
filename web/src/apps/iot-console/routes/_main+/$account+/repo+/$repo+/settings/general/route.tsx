import { CopySimple } from '~/iotconsole/components/icons';
import { useNavigate, useOutletContext, useParams } from '@remix-run/react';
import { useState } from 'react';
import { TextInput } from '~/components/atoms/input';
import { toast } from '~/components/molecule/toast';
import {
  Box,
  DeleteContainer,
} from '~/iotconsole/components/common-console-components';
import DeleteDialog from '~/iotconsole/components/delete-dialog';
import { useIotConsoleApi } from '~/iotconsole/server/gql/api-provider';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import { consoleBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { handleError } from '~/root/lib/utils/common';
import { IRepoContext } from '../../_layout';

const ProjectSettingGeneral = () => {
  const { account } = useParams();
  const { repoName } = useOutletContext<IRepoContext>();
  const api = useIotConsoleApi();
  const navigate = useNavigate();
  const [deleteRepo, setDeleteRepo] = useState(false);

  const { copy } = useClipboard({
    onSuccess() {
      toast.success('Text copied to clipboard.');
    },
  });

  return (
    <>
      <Box title="General">
        <TextInput
          label="Repo name"
          value={repoName}
          // onChange={handleChange('displayName')}
        />
        <div className="flex flex-row items-center gap-3xl">
          <div className="flex-1">
            <TextInput
              label="Repo URL"
              value={`${consoleBaseUrl}/${account}/repo/${repoName}`}
              message="This is your URL namespace within Kloudlite"
              disabled
              suffix={
                <div className="flex justify-center items-center" title="Copy">
                  <button
                    aria-label="copy"
                    onClick={() =>
                      copy(`${consoleBaseUrl}/${account}/repo/${repoName}}`)
                    }
                    className="outline-none hover:bg-surface-basic-hovered active:bg-surface-basic-active rounded text-text-default"
                    tabIndex={-1}
                  >
                    <CopySimple size={16} />
                  </button>
                </div>
              }
            />
          </div>
          <div className="flex-1">
            <TextInput
              value={repoName}
              label="Repo ID"
              message="Used when interacting with the Kloudlite API"
              suffix={
                <div className="flex justify-center items-center" title="Copy">
                  <button
                    aria-label="copy"
                    onClick={() => copy(repoName)}
                    className="outline-none hover:bg-surface-basic-hovered active:bg-surface-basic-active rounded text-text-default"
                    tabIndex={-1}
                  >
                    <CopySimple size={16} />
                  </button>
                </div>
              }
              disabled
            />
          </div>
        </div>
      </Box>
      <DeleteContainer
        title="Delete repo"
        action={async () => {
          setDeleteRepo(true);
        }}
      >
        Permanently remove your Repo and all of its contents from the Kloudlite
        platform. This action is not reversible — please continue with caution.
      </DeleteContainer>
      <DeleteDialog
        resourceName={repoName}
        resourceType="repo"
        show={deleteRepo}
        setShow={setDeleteRepo}
        onSubmit={async () => {
          try {
            const { errors } = await api.deleteRepo({
              name: repoName,
            });

            if (errors) {
              throw errors[0];
            }
            toast.success(`Repo deleted successfully`);
            setDeleteRepo(false);
            navigate(`/${account}/packages`);
          } catch (err) {
            handleError(err);
          }
        }}
      />
    </>
  );
};
export default ProjectSettingGeneral;
