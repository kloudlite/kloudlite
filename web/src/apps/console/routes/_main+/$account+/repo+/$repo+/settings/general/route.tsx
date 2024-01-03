import { CopySimple } from '@jengaicons/react';
import { useNavigate, useParams } from '@remix-run/react';
import { useState } from 'react';
import { TextInput } from '~/components/atoms/input';
import { toast } from '~/components/molecule/toast';
import {
  Box,
  DeleteContainer,
} from '~/console/components/common-console-components';
import DeleteDialog from '~/console/components/delete-dialog';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import useClipboard from '~/root/lib/client/hooks/use-clipboard';
import { consoleBaseUrl } from '~/root/lib/configs/base-url.cjs';
import { handleError } from '~/root/lib/utils/common';

const ProjectSettingGeneral = () => {
  const { repo = '', account } = useParams();
  const api = useConsoleApi();
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
          value={repo}
          // onChange={handleChange('displayName')}
        />
        <div className="flex flex-row items-center gap-3xl">
          <div className="flex-1">
            <TextInput
              label="Repo URL"
              value={`${consoleBaseUrl}/${account}/repo/${repo}`}
              message="This is your URL namespace within Kloudlite"
              disabled
              suffix={
                <div className="flex justify-center items-center" title="Copy">
                  <button
                    aria-label="copy"
                    onClick={() =>
                      copy(`${consoleBaseUrl}/${account}/repo/${repo}}`)
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
              value={repo}
              label="Repo ID"
              message="Used when interacting with the Kloudlite API"
              suffix={
                <div className="flex justify-center items-center" title="Copy">
                  <button
                    aria-label="copy"
                    onClick={() => copy(repo)}
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
        resourceName={repo}
        resourceType="repo"
        show={deleteRepo}
        setShow={setDeleteRepo}
        onSubmit={async () => {
          try {
            const { errors } = await api.deleteRepo({
              name: repo,
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
