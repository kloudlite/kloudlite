import {
  ChangeEvent,
  FormEventHandler,
  useCallback,
  useEffect,
  useState,
} from 'react';
import { Updater, useImmer } from 'use-immer';
import Yup from '../../server/helpers/yup';
import { parseError } from '../../utils/common';
import { FlatMapType } from '../../types/common';

interface useFormProps<T = any> {
  initialValues: T;
  validationSchema: any;
  onSubmit: (val: T) => any | void | Promise<any | void>;
  whileLoading?: () => void;
  disableWhileLoading?: boolean;
}

interface useFormResp<T = any> {
  values: T;
  setValues: Updater<T>;
  resetValues: () => void;
  errors: FlatMapType<string | undefined>;
  handleChange: (key: string) => (e: ChangeEvent<HTMLInputElement>) => void;
  handleSubmit: FormEventHandler<HTMLFormElement>;
  isLoading: boolean;
  submit: () => any | Promise<any>;
}
//
// type useFormType<T = any> = (props: useFormProps<T>) => useFormResp<T>;

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

  const resetValues = () => setValues(initialValues);
  const checkIsPresent = useCallback(
    async (path: string, value: any) => {
      if (errors && !errors[path]) {
        return;
      }
      // if (typeof errors === 'object' && errors !== null && !errors[path])
      //   return;

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

  const submit = async () => {
    if (values instanceof Array) {
      setErrors({});
    }
    if (isLoading && disableWhileLoading) {
      if (whileLoading) whileLoading();
      return false;
    }
    setIsLoading(true);
    try {
      await validationSchema.validate(values, {
        abortEarly: false,
      });
      try {
        const response = await onSubmit(values);
        return response;
      } catch (err) {
        console.error(err);
        // toast.error(err.message);
        return false;
        // show server error
      }
    } catch (err) {
      // show field errors
      // console.error(err);
      console.log(parseError(err).message);
      (err as Yup.ValidationError).inner.map((item) => {
        setErrors((d: any) => {
          d[item.path || ''] = item.message;
        });
        return true;
      });
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = async (e: any) => {
    // e.stopPropagation();
    e.preventDefault();
    await submit();
  };

  return {
    values,
    errors,
    handleChange,
    handleSubmit,
    setValues,
    resetValues,
    isLoading,
    submit,
  };
}

export const dummyEvent = (value: any) => {
  return { target: { value } };
};

export default useForm;
