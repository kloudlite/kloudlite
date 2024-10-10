import { useParams } from '@remix-run/react';
import { useState } from 'react';
import { Button } from '~/components/atoms/button';
import Banner from '~/components/molecule/banner';
import Popup from '~/components/molecule/popup';
import CodeView from '~/console/components/code-view';
import ExtendedFilledTab from '~/console/components/extended-filled-tab';
import { LoadingPlaceHolder } from '~/console/components/loading';
import { useConsoleApi } from '~/console/server/gql/api-provider';
import { ensureAccountClientSide } from '~/console/server/utils/auth-utils';
import useCustomSwr from '~/root/lib/client/hooks/use-custom-swr';
import { NonNullableString } from '~/root/lib/types/common';

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

  const [active, setActive] = useState<
    'url' | 'script-url' | NonNullableString
  >('script-url');

  return (
    <Popup.Root onOpenChange={onClose} show={show} className="!w-[800px]">
      <Popup.Header>Add an Image to the Registry</Popup.Header>
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
              <div className="flex flex-col gap-xl">
                <ExtendedFilledTab
                  value={active}
                  onChange={setActive}
                  items={[
                    {
                      label: 'Script',
                      to: 'script-url',
                      value: 'script-url',
                    },
                    { label: 'cURL Command', to: 'url', value: 'url' },
                  ]}
                />
                {active === 'url' && (
                  <div className="flex flex-col gap-3xl">
                    {data.url &&
                      data.url.map((u) => (
                        <CodeView
                          key={u}
                          preClassName="!overflow-none text-wrap break-words"
                          copy={false}
                          data={u}
                          language="sh"
                        />
                      ))}
                    <Banner
                      type="info"
                      body={
                        <div className="flex flex-col gap-md">
                          <span className="bodyMd-medium">
                            Webhook Image Metadata Push Example:
                          </span>
                          <span className="bodyMd">{data.urlExample}</span>
                        </div>
                      }
                    />
                  </div>
                )}
                {active === 'script-url' && (
                  <div className="flex flex-col gap-3xl">
                    {data.scriptUrl &&
                      data.scriptUrl.map((u) => (
                        <CodeView
                          key={u}
                          preClassName="!overflow-none text-wrap break-words"
                          copy={false}
                          data={u}
                          language="sh"
                        />
                      ))}
                    <Banner
                      type="info"
                      body={
                        <div className="flex flex-col gap-md">
                          <span className="bodyMd-medium">
                            Shell Script Example with Image and Meta
                            Information:
                          </span>
                          <span className="bodyMd">
                            {data.scriptUrlExample}
                          </span>
                        </div>
                      }
                    />
                  </div>
                )}
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
