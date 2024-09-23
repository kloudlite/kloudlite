import { useParams } from '@remix-run/react';
import { Button } from '~/components/atoms/button';
import Popup from '~/components/molecule/popup';
import CodeView from '~/console/components/code-view';
import { LoadingPlaceHolder } from '~/console/components/loading';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { ensureAccountClientSide } from '~/console/server/utils/auth-utils';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';

export const RegistryImageInstruction = ({
  show,
  onClose,
}: {
  show: boolean;
  onClose: () => void;
}) => {
  const params = useParams();
  ensureAccountClientSide(params);
  const api = useConsoleApi();

  const { data, isLoading, error } = useCustomSwr(
    'registry-image-instructions',
    async () => {
      return api.getRegistryImageUrl();
    }
  );

  return (
    <Popup.Root onOpenChange={onClose} show={show} className="!w-[800px]">
      <Popup.Header>Instructions to add image on registry</Popup.Header>
      <Popup.Content>
        <form className="flex flex-col gap-2xl">
          {error && (
            <span className="bodyMd-medium text-text-strong">
              Error while fetching instructions
            </span>
          )}
          {isLoading ? (
            <LoadingPlaceHolder />
          ) : (
            data && (
              <div className="flex flex-col gap-sm text-start ">
                <span className="flex flex-wrap items-center gap-md py-lg">
                  Please follow below instruction for further steps
                </span>
                <CodeView
                  preClassName="!overflow-none text-wrap break-words"
                  copy
                  data={data.url || ''}
                />

                {/* {data.url} */}
              </div>
            )
          )}
        </form>
      </Popup.Content>
      <Popup.Footer>
        <Button variant="primary-outline" content="close" onClick={onClose} />
      </Popup.Footer>
    </Popup.Root>
  );
};
