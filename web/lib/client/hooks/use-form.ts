import { useCallback, useEffect, useState } from 'react';
import { useImmer } from 'use-immer';
import Yup from '../../server/helpers/yup';
import { FlatMapType } from '../../types/common';
import { parseError } from '../../utils/common';

interface useFormProps<T = any> {
  initialValues: T;
  validationSchema: any;
  onSubmit: (val: T) => any | void | Promise<any | void>;
  whileLoading?: () => void;
  disableWhileLoading?: boolean;
}

interface useFormResp<T = any> {
  values: T;
  setValues: (fn: ((val: T) => T) | T) => void;
  resetValues: (v?: any) => void;
  errors: FlatMapType<string | undefined>;
  handleChange: (key: string) => (e: { target: { value: string } }) => void;
  handleSubmit: (e: { preventDefault: () => void }) => void;
  isLoading: boolean;
  submit: () => boolean | Promise<boolean>;
}

function useForm<T>({
  initialValues,
  validationSchema,
  onSubmit,
  whileLoading,
  disableWhileLoading = true,
}: useFormProps<T>): useFormResp<T> {
  const [values, setValues] = useImmer(initialValues);
  const [errors, setErrors] = useImmer<FlatMapType<string | undefined>>({});

  const [isLoading, setIsLoading] = useState(false);

  const resetValues = (v: any) => {
    if (v) {
      setValues(v);
    } else {
      setValues(initialValues);
    }
    setErrors({});
  };
  const checkIsPresent = useCallback(
    async (path: string, value: any) => {
      if (errors && !errors[path]) {
        return;
      }

      try {
        await validationSchema.validate(
          { ...values, [path]: value },
          {
            abortEarly: false,
          }
        );
        setErrors({});
      } catch (err) {
        const res = (err as Yup.ValidationError).inner.filter(
          (item) => item.path === path
        );
        if (res.length === 0)
          setErrors((d: any) => {
            d[path] = undefined;
          });
        else {
          setErrors((d: any) => {
            d[path] = res[0].message;
          });
        }
      }
    },
    [validationSchema, errors, setErrors, values]
  );

  useEffect(() => {
    if (Object.keys(errors).length === 0)
      Object.keys(initialValues || {}).map((key) => {
        setErrors((d: any) => {
          d[key] = undefined;
        });
        return true;
      });
  }, [initialValues, setErrors, errors]);

  const handleChange = (keyPath: string) => {
    // const ki = Yup.MixedSchema<{ name: string }>;
    // const k = Yup.object({ n: Yup.string() });
    const keyPaths = keyPath.split('.');
    if (keyPaths.length > 1) {
      return (e: any) => {
        setValues((d: any) => {
          if (
            e.target.value !== false &&
            !e.target.value &&
            e.target.value !== ''
          ) {
            delete d[keyPaths[0]][keyPaths[1]]?.[keyPaths[2]]?.[keyPaths[3]]?.[
              keyPaths[4]
            ];
          }
          if (keyPaths.length === 2) {
            d[keyPaths[0]][keyPaths[1]] = e.target.value;
          } else if (keyPaths.length === 3) {
            d[keyPaths[0]][keyPaths[1]][keyPaths[2]] = e.target.value;
          } else if (keyPaths.length === 4) {
            d[keyPaths[0]][keyPaths[1]][keyPaths[2]][keyPaths[3]] =
              e.target.value;
          }
        });
        checkIsPresent(keyPath, e.target.value);
      };
    }
    return (e: any) => {
      setValues((d: any) => {
        if (
          e.target.value !== false &&
          e.target.value !== '' &&
          !e.target.value
        ) {
          delete d[keyPath];
        } else {
          d[keyPath] = e.target.value;
        }
      });
      checkIsPresent(keyPath, e.target.value);
    };
  };

  const submit: () => Promise<boolean> = async () => {
    if (values instanceof Array) {
      setErrors({});
    }
    if (isLoading && disableWhileLoading) {
      if (whileLoading) whileLoading();
      return false;
    }
    setIsLoading(true);
    try {
      const val = await validationSchema.validate(values, {
        abortEarly: false,
      });

      try {
        await onSubmit(val);
        setIsLoading(false);
        return true;
      } catch (err) {
        console.error(err);
        setIsLoading(false);
        return false;
      }
    } catch (err) {
      console.log(parseError(err).message);
      (err as Yup.ValidationError).inner.map((item) => {
        setErrors((d: any) => {
          d[item.path || ''] = item.message;
        });
        return true;
      });
      setIsLoading(false);
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = async (e: any) => {
    e.preventDefault();
    e.stopPropagation();
    await submit();
  };

  type ISetState<T = any> = (fn: ((val: T) => T) | T) => void;
  const sv: ISetState = (fn) => {
    if (typeof fn === 'function') {
      setValues((v) => {
        return fn(v);
      });
    } else {
      setValues(fn);
    }
  };

  return {
    values,
    errors,
    handleChange,
    handleSubmit,
    setValues: sv,
    resetValues,
    isLoading,
    submit,
  };
}

export const dummyEvent = (value: any) => {
  return { target: { value } };
};

export default useForm;
