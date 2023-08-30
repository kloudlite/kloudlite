import { ArrowLineDown } from '@jengaicons/react';
import { useState } from 'react';
import { TextInput } from '~/components/atoms/input';
import Popup from '~/components/molecule/popup';
import SelectInput from '~/components/atoms/select';
import { dummyData } from '~/console/dummy/data';
import useForm from '~/root/lib/client/hooks/use-form';
import Yup from '~/root/lib/server/helpers/yup';
import { IdSelector } from '~/console/components/id-selector';

const QRPlaceholder = () => {
  return (
    <svg
      width="181"
      height="182"
      viewBox="0 0 181 182"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M110.5 68.5H105.5V81H110.5V68.5Z" fill="#111827" />
      <path d="M120.5 68.5H115.5V81H120.5V68.5Z" fill="#111827" />
      <path d="M110.5 86H105.5V91H110.5V86Z" fill="#111827" />
      <path d="M120.5 86H115.5V91H120.5V86Z" fill="#111827" />
      <path d="M93 96H88V101H93V96Z" fill="#111827" />
      <path d="M80.5 146H75.5V151H80.5V146Z" fill="#111827" />
      <path d="M120.5 96H115.5V101H120.5V96Z" fill="#111827" />
      <path d="M110.5 106H105.5V111H110.5V106Z" fill="#111827" />
      <path d="M120.5 106H115.5V111H120.5V106Z" fill="#111827" />
      <path
        d="M15.5 16H23V11H13C11.62 11 10.5 12.12 10.5 13.5V23.5H15.5V16Z"
        fill="#111827"
      />
      <path
        d="M168 11H158V16H165.5V23.5H170.5V13.5C170.5 12.12 169.38 11 168 11Z"
        fill="#111827"
      />
      <path
        d="M15.5 158.5H10.5V168.5C10.5 169.88 11.62 171 13 171H23V166H15.5V158.5Z"
        fill="#111827"
      />
      <path
        d="M165.5 166H158V171H168C169.38 171 170.5 169.88 170.5 168.5V158.5H165.5V166Z"
        fill="#111827"
      />
      <path
        d="M20.5 23.5V53.5C20.5 54.88 21.62 56 23 56H53C54.38 56 55.5 54.88 55.5 53.5V23.5C55.5 22.12 54.38 21 53 21H23C21.62 21 20.5 22.12 20.5 23.5ZM25.5 26H50.5V51H25.5V26Z"
        fill="#111827"
      />
      <path
        d="M43 31H33C31.62 31 30.5 32.12 30.5 33.5V43.5C30.5 44.88 31.62 46 33 46H43C44.38 46 45.5 44.88 45.5 43.5V33.5C45.5 32.12 44.38 31 43 31ZM40.5 41H35.5V36H40.5V41Z"
        fill="#111827"
      />
      <path
        d="M128 56H158C159.38 56 160.5 54.88 160.5 53.5V23.5C160.5 22.12 159.38 21 158 21H128C126.62 21 125.5 22.12 125.5 23.5V53.5C125.5 54.88 126.62 56 128 56ZM130.5 26H155.5V51H130.5V26Z"
        fill="#111827"
      />
      <path
        d="M148 31H138C136.62 31 135.5 32.12 135.5 33.5V43.5C135.5 44.88 136.62 46 138 46H148C149.38 46 150.5 44.88 150.5 43.5V33.5C150.5 32.12 149.38 31 148 31ZM145.5 41H140.5V36H145.5V41Z"
        fill="#111827"
      />
      <path
        d="M53 126H23C21.62 126 20.5 127.12 20.5 128.5V158.5C20.5 159.88 21.62 161 23 161H53C54.38 161 55.5 159.88 55.5 158.5V128.5C55.5 127.12 54.38 126 53 126ZM50.5 156H25.5V131H50.5V156Z"
        fill="#111827"
      />
      <path
        d="M33 151H43C44.38 151 45.5 149.88 45.5 148.5V138.5C45.5 137.12 44.38 136 43 136H33C31.62 136 30.5 137.12 30.5 138.5V148.5C30.5 149.88 31.62 151 33 151ZM35.5 141H40.5V146H35.5V141Z"
        fill="#111827"
      />
      <path
        d="M120.5 58.5H103V28.5H98V61C98 62.38 99.12 63.5 100.5 63.5H120.5V58.5Z"
        fill="#111827"
      />
      <path d="M83 41H65.5V46H83V41Z" fill="#111827" />
      <path d="M93 31H75.5V36H93V31Z" fill="#111827" />
      <path d="M25.5 58.5H20.5V81H25.5V58.5Z" fill="#111827" />
      <path
        d="M55.5 73.5H30.5V78.5H53V88.5H58V76C58 74.62 56.88 73.5 55.5 73.5Z"
        fill="#111827"
      />
      <path d="M38 61H33V66H38V61Z" fill="#111827" />
      <path d="M68 56H63V76H68V56Z" fill="#111827" />
      <path d="M85.5 81H63V86H85.5V81Z" fill="#111827" />
      <path
        d="M163 83.5H158V108.5H150.5V113.5H160.5C161.88 113.5 163 112.38 163 111V83.5Z"
        fill="#111827"
      />
      <path d="M80.5 56H75.5V61H80.5V56Z" fill="#111827" />
      <path d="M25.5 93.5H20.5V113.5H25.5V93.5Z" fill="#111827" />
      <path d="M43 116H20.5V121H43V116Z" fill="#111827" />
      <path
        d="M135.5 81H140.5V71C140.5 69.62 139.38 68.5 138 68.5H125.5V73.5H135.5V81Z"
        fill="#111827"
      />
      <path d="M38 93.5H33V98.5H38V93.5Z" fill="#111827" />
      <path d="M65.5 91H60.5V111H65.5V91Z" fill="#111827" />
      <path d="M83 116H60.5V121H83V116Z" fill="#111827" />
      <path d="M53 106H30.5V111H53V106Z" fill="#111827" />
      <path d="M68 141H63V151H68V141Z" fill="#111827" />
      <path d="M83 131H60.5V136H83V131Z" fill="#111827" />
      <path
        d="M100.5 121V108.5C100.5 107.12 99.38 106 98 106H73V111H95.5V121H100.5Z"
        fill="#111827"
      />
      <path d="M80.5 91H75.5V96H80.5V91Z" fill="#111827" />
      <path d="M53 96H48V101H53V96Z" fill="#111827" />
      <path d="M163 143.5H158V161H163V143.5Z" fill="#111827" />
      <path d="M153 156H145.5V161H153V156Z" fill="#111827" />
      <path d="M150.5 141H145.5V151H150.5V141Z" fill="#111827" />
      <path
        d="M163 121C163 119.62 161.88 118.5 160.5 118.5H133V123.5H158V138.5H163V121Z"
        fill="#111827"
      />
      <path
        d="M130.5 151V138.5C130.5 137.12 129.38 136 128 136H100.5V141H125.5V151H130.5Z"
        fill="#111827"
      />
      <path d="M153 128.5H135.5V133.5H153V128.5Z" fill="#111827" />
      <path d="M78 156H60.5V161H78V156Z" fill="#111827" />
      <path d="M90.5 138.5H85.5V161H90.5V138.5Z" fill="#111827" />
      <path d="M100.5 126H90.5V131H100.5V126Z" fill="#111827" />
      <path
        d="M128 101H140.5V96H130.5V81H125.5V98.5C125.5 99.88 126.62 101 128 101Z"
        fill="#111827"
      />
      <path d="M148 86H135.5V91H148V86Z" fill="#111827" />
      <path d="M153 61H145.5V66H153V61Z" fill="#111827" />
      <path d="M163 61H158V73.5H163V61Z" fill="#111827" />
      <path d="M150.5 73.5H145.5V78.5H150.5V73.5Z" fill="#111827" />
      <path
        d="M108 53.5H118C119.38 53.5 120.5 52.38 120.5 51V18.5H115.5V48.5H108V53.5Z"
        fill="#111827"
      />
      <path
        d="M70.5 23.5H108V18.5H68C66.62 18.5 65.5 19.62 65.5 21V33.5H70.5V23.5Z"
        fill="#111827"
      />
      <path
        d="M73 71H90.5C91.88 71 93 69.88 93 68.5V48.5H88V66H73V71Z"
        fill="#111827"
      />
      <path d="M100.5 68.5H95.5V78.5H100.5V68.5Z" fill="#111827" />
      <path d="M118 146H100.5V151H118V146Z" fill="#111827" />
      <path d="M123 156H100.5V161H123V156Z" fill="#111827" />
      <path
        d="M135.5 156H130.5V161H138C139.38 161 140.5 159.88 140.5 158.5V138.5H135.5V156Z"
        fill="#111827"
      />
      <path d="M55.5 116H50.5V121H55.5V116Z" fill="#111827" />
      <path d="M118 116H108V121H118V116Z" fill="#111827" />
      <path d="M130.5 106H125.5V113.5H130.5V106Z" fill="#111827" />
      <path d="M128 118.5H123V123.5H128V118.5Z" fill="#111827" />
      <path d="M115.5 126H108V131H115.5V126Z" fill="#111827" />
      <path d="M140.5 106H135.5V111H140.5V106Z" fill="#111827" />
      <path d="M153 96H145.5V101H153V96Z" fill="#111827" />
      <path d="M110.5 96H98V101H110.5V96Z" fill="#111827" />
      <path d="M100.5 83.5H95.5V91H100.5V83.5Z" fill="#111827" />
      <rect x="0.5" y="1" width="180" height="180" rx="8" stroke="#D4D4D8" />
    </svg>
  );
};

export const ShowQR = ({ show, setShow, data = {} }) => {
  return (
    <Popup.Root show={show} onOpenChange={setShow}>
      <Popup.Header>QR Code</Popup.Header>
      <Popup.Content>
        <div className="flex flex-row gap-7xl">
          <div className="flex flex-col gap-2xl">
            <div className="bodyLg-medium text-text-default">
              Use WireGuard on your phone
            </div>
            <ul className="flex flex-col gap-lg bodyMd text-text-default list-disc list-outside pl-2xl">
              <li>Download the app from Google Play or Apple Store</li>
              <li>Open the app on your Phone</li>
              <li>Tab on the ➕ Plus icon</li>
              <li>Point your phone to this screen to capture the QR code</li>
            </ul>
          </div>
          <div>
            <QRPlaceholder />
          </div>
        </div>
      </Popup.Content>
    </Popup.Root>
  );
};

export const ShowWireguardConfig = ({ show, setShow, data = {} }) => {
  return (
    <Popup.Root show={show} onOpenChange={setShow}>
      <Popup.Header>WireGuard Config</Popup.Header>
      <Popup.Content>
        <div className="flex flex-col gap-3xl">
          <div className="bodyMd text-text-default">
            Please use the following configuration to set up your WireGuard
            client.
          </div>
          <div className="p-3xl flex flex-col gap-3xl border border-border-default rounded-lg">
            <div className="pb-3xl flex flex-col gap-lg">
              <div className="bodyMd-medium text-text-soft">Interface</div>
              <div className="flex flex-col gap-md text-text-default">
                <div className="flex flex-row gap-4xl ">
                  <span className="bodyMd-medium w-9xl">PrivateKey</span>
                  <span className="bodyMd w-[7px]">-</span>
                  <span className="bodyMd">YJGz9Lk/80Q</span>
                </div>
                <div className="flex flex-row gap-4xl">
                  <span className="bodyMd-medium w-9xl">Address</span>
                  <span className="bodyMd w-[7px]">-</span>
                  <span className="bodyMd">10.6.0.2/32</span>
                </div>
              </div>
            </div>
            <div className="flex flex-col gap-lg">
              <div className="bodyMd-medium text-text-soft">Peer</div>
              <div className="flex flex-col gap-md text-text-default">
                <div className="flex flex-row gap-4xl">
                  <span className="bodyMd-medium w-9xl">PublicKey</span>
                  <span className="bodyMd w-[7px]">-</span>
                  <span className="bodyMd">Yy4QH9ik6vbl</span>
                </div>
                <div className="flex flex-row gap-4xl">
                  <span className="bodyMd-medium w-9xl">AllowedIPs</span>
                  <span className="bodyMd w-[7px]">-</span>
                  <span className="bodyMd">0.0.0.0/0</span>
                </div>
                <div className="flex flex-row gap-4xl">
                  <span className="bodyMd-medium w-9xl">Endpoint</span>
                  <span className="bodyMd w-[7px]">-</span>
                  <span className="bodyMd">PersistentKeepalive/25</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Popup.Content>
      <Popup.Footer>
        <Popup.Button
          content="Export"
          prefix={<ArrowLineDown />}
          variant="primary"
        />
      </Popup.Footer>
    </Popup.Root>
  );
};

const HandleDevice = ({ show, setShow }) => {
  const [clusters, _setProvisionTypes] = useState(dummyData.clusterList);

  const { values, errors, handleChange, handleSubmit, resetValues } = useForm({
    initialValues: {
      name: '',
      cluster: '',
    },
    validationSchema: Yup.object({}),
    onSubmit: async () => {},
  });

  return (
    <Popup.Root
      show={show}
      onOpenChange={(e) => {
        if (!e) {
          resetValues();
        }

        setShow(e);
      }}
    >
      <Popup.Header>
        {show.type === 'add' ? 'Add new device' : 'Edit device'}
      </Popup.Header>
      <form onSubmit={handleSubmit}>
        <Popup.Content>
          <div className="flex flex-col gap-2xl">
            <TextInput
              label="Name"
              value={values.name}
              onChange={handleChange('name')}
            />
            {show.type === 'add' && (
              <IdSelector
                name={values.name}
                onChange={(value) => handleChange('id')({ target: { value } })}
              />
            )}
            {show.type === 'add' && (
              <SelectInput.Root
                value={values.cluster}
                label="Cluster"
                onChange={(value) =>
                  handleChange('name')({ target: { value } })
                }
              >
                <SelectInput.Option disabled value="">
                  --Select--
                </SelectInput.Option>
                {clusters.map((pt) => (
                  <SelectInput.Option value={pt.value} key={pt.id}>
                    {pt.name}
                  </SelectInput.Option>
                ))}
              </SelectInput.Root>
            )}
          </div>
        </Popup.Content>
        <Popup.Footer>
          <Popup.Button closable content="Cancel" variant="basic" />
          <Popup.Button
            type="submit"
            content={show.type === 'add' ? 'Create' : 'Update'}
            variant="primary"
          />
        </Popup.Footer>
      </form>
    </Popup.Root>
  );
};

const _Wrapper = ({ show, setShow }) => {
  if (show) {
    return <HandleDevice show={show} setShow={setShow} />;
  }
  return null;
};

export default _Wrapper;
