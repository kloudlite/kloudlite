import Yup from '../../server/helpers/yup';

export const nameConv = {
  regex: /^[a-z][\w-]{1,61}[\w]$/,
  msg: 'name must be less than 63 character',
};

export const nameRule = Yup.string()
  .trim()
  .required('value is required')
  .min(3, 'must be more than 2 character')
  .max(63, 'must be lessthan 64 character')
  .test('kube-name-start', 'must start with an alphabetic character', (val) => {
    if (val.length === 0) return false;
    return /^[a-zA-Z]$/.test(val[0]);
  })
  .test('kube-name-end', 'end with an alphanumeric character', (val) => {
    if (val.length === 0) return false;
    return /^[a-zA-Z0-9]$/.test(val[val.length - 1]);
  })
  .matches(
    /^[a-z][\w-]{1,61}[\w]$/,
    "contain only lowercase alphanumeric characters or '-'"
  );
