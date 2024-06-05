/* eslint-disable guard-for-in */
import {
  AWSlogoFill,
  ChevronRight,
  GoogleCloudlogo,
} from '~/console/components/icons';
import { Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider as CloudProviders } from '~/root/src/generated/gql/server';
import { cn } from '~/components/utils';
import yup from '~/root/lib/server/helpers/yup';
import {
  IMSvTemplate,
  IMSvTemplates,
} from '../server/gql/queries/managed-templates-queries';
import { parseValue } from '../page-components/util';

export const getManagedTemplate = ({
  templates,
  kind,
  apiVersion,
}: {
  templates: IMSvTemplates;
  kind: string;
  apiVersion: string;
}): IMSvTemplate | undefined => {
  return templates
    ?.flatMap((t) => t.items.flat())
    .find((t) => t.kind === kind && t.apiVersion === apiVersion);
};

export const DIALOG_TYPE = Object.freeze({
  ADD: 'add',
  EDIT: 'edit',
  NONE: 'none',
});

export const DIALOG_DATA_NONE = Object.freeze({
  type: DIALOG_TYPE.NONE,
  data: null,
});

export const ACCOUNT_ROLES = Object.freeze({
  account_member: 'Member',
  account_admin: 'Admin',
});

interface IPopupWindowOptions {
  url: string;
  width?: number;
  height?: number;
  title?: string;
}

export const popupWindow = ({
  url,
  onClose = () => {},
  width = 800,
  height = 500,
  title = 'kloudlite',
}: IPopupWindowOptions & {
  onClose?: () => void;
}) => {
  const frame = window.open(
    url,
    title,
    `toolbar=no,scrollbars=yes,resizable=no,top=${
      window.screen.height / 2 - height / 2
    },left=${window.screen.width / 2 - width / 2},width=800,height=600`
  );

  const interval = setInterval(() => {
    if (frame && frame.closed) {
      clearInterval(interval);
      onClose();
    }
  }, 100);
};

export const asyncPopupWindow = (options: IPopupWindowOptions) => {
  return new Promise((resolve) => {
    popupWindow({
      ...options,
      onClose: () => {
        resolve(true);
      },
    });
  });
};

export const downloadFile = ({
  filename,
  data,
  format,
}: {
  filename: string;
  format: string;
  data: string;
}) => {
  const blob = new Blob([data], { type: format });

  const url = URL.createObjectURL(blob);

  const link = document.createElement('a');
  link.href = url;
  link.download = filename;

  document.body.appendChild(link);

  link.click();

  URL.revokeObjectURL(url);
  document.body.removeChild(link);
};

export const providerIcons = (iconsSize = 16) => {
  return {
    aws: <AWSlogoFill className="inline" size={iconsSize} />,
    gcp: <GoogleCloudlogo className="inline" size={iconsSize} />,
  };
};

export const renderCloudProvider = ({
  cloudprovider,
}: {
  cloudprovider: CloudProviders | 'unknown';
}) => {
  const iconSize = 16;
  switch (cloudprovider) {
    case 'aws':
      return (
        <span>
          {providerIcons(iconSize).aws}
          <span className="pl-lg bodySm-semibold">{cloudprovider}</span>
        </span>
      );
    case 'gcp':
      return (
        <span>
          {providerIcons(iconSize).gcp}
          <span className="pl-lg bodyMd-semibold">{cloudprovider}</span>
        </span>
      );
    default:
      return cloudprovider;
  }
};

// export const flatMap = (data: any) => {
//   const keys = data.split('.');
//
//   const jsonObject = keys.reduceRight(
//     (acc: any, key: string) => ({ [key]: acc }),
//     null
//   );
//   return jsonObject;
// };

export const tabIconSize = 16;
export const breadcrumIconSize = 14;
export const BreadcrumChevronRight = () => (
  <span className="text-icon-disabled">
    <ChevronRight size={breadcrumIconSize} />
  </span>
);

export const BreadcrumSlash = () => (
  <span className="text-text-disabled font-light">/</span>
);

export const BreadcrumButtonContent = ({
  content,
  className,
}: {
  content: string;
  className?: string;
}) => (
  <div className={cn('flex flex-row items-center', className)}>
    <span className="">{content}</span>
  </div>
);

export const flatM = (
  obj: Record<
    string,
    {
      defaultValue: number | string | boolean;
      inputType: string;
      multiplier?: number;
      unit?: string;
    }
  >
) => {
  const flatJson = {};
  for (const key in obj) {
    const parts = key.split('.');

    let temp: Record<string, any> = flatJson;

    if (parts.length === 1) {
      temp[key] = null;
    } else {
      parts.forEach((part, index) => {
        temp[part] = temp[part] || {};

        if (index === parts.length - 1) {
          temp[part] = obj[key].defaultValue + (obj[key].unit || '');

          if (
            typeof obj[key].defaultValue === 'number' ||
            typeof obj[key].defaultValue === 'bigint'
          ) {
            temp[part] =
              Number(obj[key].defaultValue) * (obj[key].multiplier || 1) +
              (obj[key].unit || '');
          }

          if (obj[key].inputType === 'Resource') {
            temp[part] = {
              min:
                Number(obj[key].defaultValue) * (obj[key].multiplier || 1) +
                (obj[key].unit || ''),
              max:
                Number(obj[key].defaultValue) * (obj[key].multiplier || 1) +
                (obj[key].unit || ''),
            };
          }
        }

        temp = temp[part];
      });
    }
  }

  return flatJson;
};

// const customMinMaxNumber = (min: number, max: number) => {
//   return yup
//     .string()
//     .required()
//     .test(
//       'is-valid-number',
//       `Number must be between ${min} and ${max}`,
//       (value) => {
//         // Extract the number from the string
//         const resp = parseValue(value, -1);
//         if (resp === -1) {
//           return false;
//         }
//
//         return resp >= min && resp <= max;
//       }
//     );
// };

const customMinNumber = (min: number) => {
  return yup
    .string()
    .required()
    .test('is-valid-number', `Number must be greater than ${min}`, (value) => {
      // Extract the number from the string
      const resp = parseValue(value, -1);
      if (resp === -1) {
        return false;
      }

      return resp >= min;
    });
};

const customMaxNumber = (max: number) => {
  return yup
    .string()
    .required()
    .test('is-valid-number', `Number must be less than ${max}`, (value) => {
      // Extract the number from the string
      const resp = parseValue(value, -1);
      if (resp === -1) {
        return false;
      }

      return resp <= max;
    });
};

export const flatMapValidations = (obj: Record<string, any>) => {
  const flatJson = {};
  for (const key in obj) {
    const parts = key.split('.');
    const temp: Record<string, any> = flatJson;
    // console.log('validations', obj[key]);
    if (parts.length === 1) {
      temp[key] = (() => {
        let returnYup;
        switch (obj[key].inputType) {
          case 'Number':
            returnYup = yup.number().required();

            if (obj[key].min)
              returnYup = customMinNumber(
                obj[key].min * (obj[key].multiplier || 1)
              );

            if (obj[key].max)
              returnYup = customMaxNumber(
                obj[key].max * (obj[key].multiplier || 1)
              );
            break;

          case 'String':
            returnYup = yup.string();
            break;
          case 'Resource':
            returnYup = yup.object({
              min: customMinNumber(obj[key].min * (obj[key].multiplier || 1)),
              max: customMaxNumber(obj[key].max * (obj[key].multiplier || 1)),
            });
            break;

          default:
            throw new Error('Invalid input type');
        }

        if (obj[key].required) {
          returnYup = returnYup.required();
        }

        return returnYup;
      })();
    } else {
      temp[parts[0]] = (() => {
        let resp = yup.object(
          flatMapValidations({
            [parts.slice(1, parts.length).join('.')]: obj[key],
          })
        );
        if (temp[parts[0]]) {
          resp = resp.concat(temp[parts[0]]);
        }

        return resp;
      })();
    }
  }

  return flatJson;
};
