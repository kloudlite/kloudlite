var __assign =
  (this && this.__assign) ||
  function () {
    __assign =
      Object.assign ||
      function (t) {
        for (var s, i = 1, n = arguments.length; i < n; i++) {
          s = arguments[i];
          for (const p in s)
            if (Object.prototype.hasOwnProperty.call(s, p)) t[p] = s[p];
        }
        return t;
      };
    return __assign.apply(this, arguments);
  };
Object.defineProperty(exports, '__esModule', { value: true });
exports.flatMapValidations =
  exports.flatM =
  exports.BreadcrumButtonContent =
  exports.BreadcrumSlash =
  exports.BreadcrumChevronRight =
  exports.breadcrumIconSize =
  exports.tabIconSize =
  exports.renderCloudProvider =
  exports.providerIcons =
  exports.downloadFile =
  exports.asyncPopupWindow =
  exports.popupWindow =
  exports.ACCOUNT_ROLES =
  exports.DIALOG_DATA_NONE =
  exports.DIALOG_TYPE =
    void 0;
/* eslint-disable guard-for-in */
const icons_1 = require('~/iotconsole/components/icons');
const utils_1 = require('~/components/utils');
const yup_1 = require('~/root/lib/server/helpers/yup');
const util_1 = require('../page-components/util');

exports.DIALOG_TYPE = Object.freeze({
  ADD: 'add',
  EDIT: 'edit',
  NONE: 'none',
});
exports.DIALOG_DATA_NONE = Object.freeze({
  type: exports.DIALOG_TYPE.NONE,
  data: null,
});
exports.ACCOUNT_ROLES = Object.freeze({
  account_member: 'Member',
  account_admin: 'Admin',
});
const popupWindow = function (_a) {
  const { url } = _a;
  const _b = _a.onClose;
  const onClose = _b === void 0 ? function () {} : _b;
  const _c = _a.width;
  const width = _c === void 0 ? 800 : _c;
  const _d = _a.height;
  const height = _d === void 0 ? 500 : _d;
  const _e = _a.title;
  const title = _e === void 0 ? 'kloudlite' : _e;
  const frame = window.open(
    url,
    title,
    'toolbar=no,scrollbars=yes,resizable=no,top='
      .concat(window.screen.height / 2 - height / 2, ',left=')
      .concat(window.screen.width / 2 - width / 2, ',width=800,height=600')
  );
  var interval = setInterval(function () {
    if (frame && frame.closed) {
      clearInterval(interval);
      onClose();
    }
  }, 100);
};
exports.popupWindow = popupWindow;
const asyncPopupWindow = function (options) {
  return new Promise(function (resolve) {
    (0, exports.popupWindow)({
      ...options,
      onClose() {
        resolve(true);
      },
    });
  });
};
exports.asyncPopupWindow = asyncPopupWindow;
const downloadFile = function (_a) {
  const { filename } = _a;
  const { data } = _a;
  const { format } = _a;
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
exports.downloadFile = downloadFile;
const providerIcons = function (iconsSize) {
  if (iconsSize === void 0) {
    iconsSize = 16;
  }
  return {
    aws: <icons_1.AWSlogoFill size={iconsSize} />,
    gcp: <icons_1.GoogleCloudlogo size={iconsSize} />,
  };
};
exports.providerIcons = providerIcons;
const renderCloudProvider = function (_a) {
  const { cloudprovider } = _a;
  const iconSize = 16;
  switch (cloudprovider) {
    case 'aws':
      return (
        <div className="flex flex-row gap-xl items-center">
          {(0, exports.providerIcons)(iconSize).aws}
          <span>{cloudprovider}</span>
        </div>
      );
    case 'gcp':
      return (
        <div className="flex flex-row gap-xl items-center">
          {(0, exports.providerIcons)(iconSize).gcp}
          <span>{cloudprovider}</span>
        </div>
      );
    default:
      return cloudprovider;
  }
};
exports.renderCloudProvider = renderCloudProvider;
// export const flatMap = (data: any) => {
//   const keys = data.split('.');
//
//   const jsonObject = keys.reduceRight(
//     (acc: any, key: string) => ({ [key]: acc }),
//     null
//   );
//   return jsonObject;
// };
exports.tabIconSize = 16;
exports.breadcrumIconSize = 14;
const BreadcrumChevronRight = function () {
  return (
    <span className="text-icon-disabled">
      <icons_1.ChevronRight size={exports.breadcrumIconSize} />
    </span>
  );
};
exports.BreadcrumChevronRight = BreadcrumChevronRight;
const BreadcrumSlash = function () {
  return <span className="text-text-disabled font-light">/</span>;
};
exports.BreadcrumSlash = BreadcrumSlash;
const BreadcrumButtonContent = function (_a) {
  const { content } = _a;
  const { className } = _a;
  return (
    <div className={(0, utils_1.cn)('flex flex-row items-center', className)}>
      <span className="">{content}</span>
    </div>
  );
};
exports.BreadcrumButtonContent = BreadcrumButtonContent;
const flatM = function (obj) {
  const flatJson = {};
  const _loop_1 = function (key) {
    const parts = key.split('.');
    let temp = flatJson;
    if (parts.length === 1) {
      temp[key] = null;
    } else {
      parts.forEach(function (part, index) {
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
  };
  for (const key in obj) {
    _loop_1(key);
  }
  return flatJson;
};
exports.flatM = flatM;
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
const customMinNumber = function (min) {
  return yup_1.default
    .string()
    .required()
    .test(
      'is-valid-number',
      'Number must be greater than '.concat(min),
      function (value) {
        // Extract the number from the string
        const resp = (0, util_1.parseValue)(value, -1);
        if (resp === -1) {
          return false;
        }
        return resp >= min;
      }
    );
};
const customMaxNumber = function (max) {
  return yup_1.default
    .string()
    .required()
    .test(
      'is-valid-number',
      'Number must be less than '.concat(max),
      function (value) {
        // Extract the number from the string
        const resp = (0, util_1.parseValue)(value, -1);
        if (resp === -1) {
          return false;
        }
        return resp <= max;
      }
    );
};
const flatMapValidations = function (obj) {
  const flatJson = {};
  const _loop_2 = function (key) {
    const parts = key.split('.');
    const temp = flatJson;
    // console.log('validations', obj[key]);
    if (parts.length === 1) {
      temp[key] = (function () {
        let returnYup;
        switch (obj[key].inputType) {
          case 'Number':
            returnYup = yup_1.default.number().required();
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
            returnYup = yup_1.default.string();
            break;
          case 'Resource':
            returnYup = yup_1.default.object({
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
      temp[parts[0]] = (function () {
        let _a;
        let resp = yup_1.default.object(
          (0, exports.flatMapValidations)(
            ((_a = {}),
            (_a[parts.slice(1, parts.length).join('.')] = obj[key]),
            _a)
          )
        );
        if (temp[parts[0]]) {
          resp = resp.concat(temp[parts[0]]);
        }
        return resp;
      })();
    }
  };
  for (const key in obj) {
    _loop_2(key);
  }
  return flatJson;
};
exports.flatMapValidations = flatMapValidations;
