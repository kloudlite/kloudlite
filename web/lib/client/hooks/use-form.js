import { useCallback, useEffect, useState } from 'react';
import { useImmer } from 'use-immer';

const defFun = async (_) => {};
function useForm({
  initialValues,
  validationSchema,
  onSubmit = defFun,
  whileLoading = defFun,
  disableWhileLoading = true,
}) {
  const [values, setValues] = useImmer(initialValues);
  const [errors, setErrors] = useImmer({});
  const [isLoading, setIsLoading] = useState(false);

  const resetValues = () => setValues(initialValues);
  const checkIsPresent = useCallback(
    async (path, value) => {
      if (!errors[path]) return;

      try {
        await validationSchema.validate(
          { ...values, [path]: value },
          {
            abortEarly: false,
          }
        );
        setErrors({});
      } catch (err) {
        const res = err.inner.filter((item) => item.path === path);
        if (res.length === 0)
          setErrors((d) => {
            d[path] = undefined;
          });
        else {
          setErrors((d) => {
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
        setErrors((d) => {
          d[key] = undefined;
        });
        return true;
      });
  }, [initialValues, setErrors, errors]);

  const handleChange = (keyPath) => {
    const keyPaths = keyPath.split('.');
    if (keyPaths.length > 1) {
      return (e) => {
        setValues((d) => {
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
    return (e) => {
      setValues((d) => {
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
      whileLoading();
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
      console.log(err.message);
      err.inner.map((item) => {
        setErrors((d) => {
          d[item.path] = item.message;
        });
        return true;
      });
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = async (e) => {
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

export default useForm;
